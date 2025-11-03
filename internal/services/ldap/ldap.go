// Package ldap provides functionality for LDAP/Active Directory integration
//
//nolint:gocyclo // need refactoring
package ldap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rom8726/di"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

var _ di.Servicer = (*Service)(nil)

var (
	ErrLDAPDisabled       = errors.New("ldap integration is disabled")
	ErrLDAPNotConfigured  = errors.New("ldap is not properly configured")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ClientFactory func(config *ClientConfig) (ClientService, error)

type Config struct {
	Enabled         bool
	URL             string
	BindDN          string
	BindPassword    string
	UserBaseDN      string
	UserFilter      string
	UserNameAttr    string
	UserEmailAttr   string
	GroupBaseDN     string
	GroupFilter     string
	GroupNameAttr   string
	GroupMemberAttr string
	StartTLS        bool
	InsecureTLS     bool
	Timeout         time.Duration
	SyncInterval    time.Duration
}

type Service struct {
	enabled           bool
	currentConfig     *domain.LDAPConfig
	client            ClientService
	clientFactory     ClientFactory
	userRepo          contract.UsersRepository
	ldapSyncLogsRepo  contract.LDAPSyncLogsRepository
	ldapSyncStatsRepo contract.LDAPSyncStatsRepository
	settingsService   contract.SettingsUseCase
	mu                sync.RWMutex
	syncInterval      time.Duration

	ctx       context.Context //nolint:containedctx // it's ok here
	ctxCancel context.CancelFunc

	// Sync status tracking
	syncStatus    domain.LDAPSyncStatus
	syncProgress  domain.LDAPSyncProgress
	syncMutex     sync.RWMutex
	syncCancelCtx context.CancelFunc
}

func New(
	userRepo contract.UsersRepository,
	ldapSyncLogsRepo contract.LDAPSyncLogsRepository,
	settingsService contract.SettingsUseCase,
	ldapSyncStatsRepo contract.LDAPSyncStatsRepository,
) (*Service, error) {
	service := &Service{
		userRepo:          userRepo,
		ldapSyncLogsRepo:  ldapSyncLogsRepo,
		settingsService:   settingsService,
		ldapSyncStatsRepo: ldapSyncStatsRepo,
		clientFactory:     func(config *ClientConfig) (ClientService, error) { return NewClient(config) },
	}

	if err := service.reloadConfig(context.TODO(), "init service"); err != nil {
		return nil, fmt.Errorf("failed to reload LDAP config: %w", err)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.ctxCancel = ctxCancel

	go service.StartSyncJob(context.TODO())

	return service, nil
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (bool, error) {
	// Reload config before authentication
	if err := s.reloadConfig(ctx, "authenticate"); err != nil {
		return false, fmt.Errorf("failed to reload LDAP config: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isEnabled() {
		return false, ErrLDAPDisabled
	}

	if s.client == nil {
		return false, ErrLDAPNotConfigured
	}

	authenticated, err := s.client.Authenticate(ctx, username, password)
	if err != nil {
		return false, fmt.Errorf("LDAP authentication failed: %w", err)
	}

	if !authenticated {
		return false, ErrInvalidCredentials
	}

	// Ensure the user exists locally
	_, err = s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if err := s.syncUser(ctx, username); err != nil {
			return false, fmt.Errorf("failed to sync user: %w", err)
		}
	}

	return true, nil
}

// TestConnection tests the connection to the LDAP server using current configuration.
func (s *Service) TestConnection(ctx context.Context) error {
	// Reload config before testing
	if err := s.reloadConfig(ctx, "test connection"); err != nil {
		return fmt.Errorf("failed to reload LDAP config: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isEnabled() {
		return ErrLDAPDisabled
	}

	if s.client == nil {
		return ErrLDAPNotConfigured
	}

	return s.client.TestConnection(ctx)
}

// Close closes the LDAP client connection.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		return s.client.Close()
	}

	return nil
}

func (s *Service) Start(context.Context) error {
	// No longer starting sync job automatically at startup
	// Sync will be triggered manually through the admin UI
	return nil
}

func (s *Service) Stop(context.Context) error {
	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	return nil
}

// syncUser synchronizes a single user from LDAP to the local database.
func (s *Service) syncUser(ctx context.Context, username string) error {
	// Get user details from LDAP
	userAttrs, err := s.client.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user from LDAP: %w", err)
	}

	// Extract user attributes
	config, err := s.settingsService.GetLDAPConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LDAP config: %w", err)
	}

	email := ""
	if emails, emailOk := userAttrs[config.UserEmailAttr]; emailOk && len(emails) > 0 {
		email = emails[0]
	}

	// Check if a user exists
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	// Create or update a user
	if errors.Is(err, domain.ErrEntityNotFound) {
		// Create a new user
		userDTO := domain.UserDTO{
			Username:      username,
			Email:         email,
			IsSuperuser:   false,
			PasswordHash:  "", // No password needed for LDAP users
			IsTmpPassword: false,
			IsExternal:    true, // Mark as external user from LDAP
		}

		user, err = s.userRepo.Create(ctx, userDTO)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user
		user.Username = username
		user.Email = email
		user.IsActive = true
		user.IsSuperuser = false
		user.IsExternal = true // Mark as an external user from LDAP

		if err := s.userRepo.Update(ctx, &user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	}

	return nil
}

// SyncUsers synchronizes all users from LDAP to the local database.
func (s *Service) SyncUsers(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isEnabled() || s.client == nil {
		return ErrLDAPNotConfigured
	}

	// Get all users from LDAP
	users, err := s.client.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users from LDAP: %w", err)
	}

	// Sync each user
	for _, username := range users {
		if err := s.syncUser(ctx, username); err != nil {
			// Log error but continue with other users
			slog.Error("failed to sync user", "username", username, "error", err)
		}
	}

	return nil
}

// StartSyncJob starts a background job to periodically sync users and groups from LDAP.
func (s *Service) StartSyncJob(ctx context.Context) {
	if !s.isEnabled() || s.client == nil {
		slog.Warn("LDAP integration is disabled or not configured, skipping sync job")

		return
	}

	// Start periodic sync if an interval is set
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if s.syncInterval <= 0 {
			select {
			case <-s.ctx.Done():
				continue
			case <-time.After(time.Second * 10):
			}
		}

		s.backgroundSync(s.ctx) //nolint:contextcheck // false positive
	}
}

func (s *Service) ReloadConfig(ctx context.Context) error {
	return s.reloadConfig(ctx, "user triggered action")
}

func (s *Service) backgroundSync(ctx context.Context) {
	if s.syncInterval <= 0 {
		return
	}

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			syncID := uuid.NewString()

			s.syncMutex.Lock()
			s.syncStatus.IsRunning = true
			s.syncMutex.Unlock()

			// Create log entry
			log := domain.LDAPSyncLog{
				Timestamp:     time.Now(),
				Level:         "info",
				Message:       "Background sync job started",
				SyncSessionID: syncID,
			}
			// Write log to a database
			if _, logErr := s.ldapSyncLogsRepo.Create(ctx, log); logErr != nil {
				slog.Error("Failed to write LDAP sync log", "error", logErr, "syncID", syncID)
			}

			slog.Info("Start syncing users and groups from LDAP", "syncID", syncID)

			var firstErr error

			if err := s.SyncUsers(ctx); err != nil {
				// Log error but continue
				slog.Error("Error syncing users from LDAP", "error", err)

				if firstErr == nil {
					firstErr = err
				}
			}

			s.syncMutex.Lock()
			s.syncStatus.IsRunning = false
			s.syncMutex.Unlock()

			level := "info"
			message := "Background sync job completed"

			if firstErr != nil {
				level = "error"
				message = "Error syncing users and groups from LDAP"
			}

			// Create log entry
			log = domain.LDAPSyncLog{
				Timestamp:     time.Now(),
				Level:         level,
				Message:       message,
				SyncSessionID: syncID,
			}
			// Write log to a database
			if _, logErr := s.ldapSyncLogsRepo.Create(ctx, log); logErr != nil {
				slog.Error("Failed to write LDAP sync log", "error", logErr, "syncID", syncID)
			}

		case <-ctx.Done():
			return
		}
	}
}

// reloadConfig reloads LDAP configuration from settings.
//
//nolint:gosec // it's ok number conversations
func (s *Service) reloadConfig(ctx context.Context, reason string) error {
	config, err := s.settingsService.GetLDAPConfig(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			slog.Info("LDAP integration is disabled, skipping config reload", "reason", reason)
			s.mu.Lock()
			s.enabled = false
			s.mu.Unlock()

			return nil
		}

		return fmt.Errorf("failed to get LDAP config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	firstStart := s.currentConfig == nil

	s.enabled = config.Enabled

	if !s.enabled {
		s.client = nil

		return nil
	}

	// Parse timeout
	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		timeout = 30 * time.Second // default timeout
	}

	s.syncInterval = time.Second * time.Duration(config.SyncInterval)

	if !firstStart && *s.currentConfig == *config {
		return nil
	}

	s.currentConfig = config

	slog.Info("Create new LDAP client. May take a few seconds...",
		"url", config.URL, "timeout", timeout, "reason", reason)

	client, err := s.clientFactory(&ClientConfig{
		URL:           config.URL,
		BindDN:        config.BindDN,
		BindPassword:  config.BindPassword,
		UserBaseDN:    config.UserBaseDN,
		UserFilter:    config.UserFilter,
		UserNameAttr:  config.UserNameAttr,
		UserEmailAttr: config.UserEmailAttr,
		StartTLS:      config.StartTLS,
		InsecureTLS:   config.InsecureTLS,
		Timeout:       timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to create LDAP client: %w", err)
	}

	// Close existing client if any
	if s.client != nil {
		_ = s.client.Close()
	}

	s.client = client

	if s.ctxCancel != nil {
		s.ctxCancel()
		time.Sleep(time.Millisecond * 100)

		s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	}

	return nil
}

// isEnabled checks if LDAP is enabled both in configuration and by license.
func (s *Service) isEnabled() bool {
	return s.enabled
}
