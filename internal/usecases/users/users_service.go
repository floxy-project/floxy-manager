package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/passworder"
)

type UsersService struct {
	usersRepo        contract.UsersRepository
	tokenizer        contract.Tokenizer
	emailer          contract.Emailer
	twoFARateLimiter contract.TwoFARateLimiter
	ssoManager       contract.SSOProviderManager
	authProvider     AuthProvider
}

func New(
	usersRepo contract.UsersRepository,
	tokenizer contract.Tokenizer,
	emailer contract.Emailer,
	twoFARateLimiter contract.TwoFARateLimiter,
	ssoManager contract.SSOProviderManager,
	authProviders []AuthProvider,
) *UsersService {
	// Create a chain of authentication providers
	authProvider := NewAuthProviderChain(
		// Add all authentication providers
		authProviders...,
	)

	// Add a local auth provider as the last resort
	localAuthProvider := NewLocalAuthProvider(usersRepo)
	authProvider.providers = append(authProvider.providers, localAuthProvider)

	return &UsersService{
		usersRepo:        usersRepo,
		tokenizer:        tokenizer,
		emailer:          emailer,
		twoFARateLimiter: twoFARateLimiter,
		authProvider:     authProvider,
		ssoManager:       ssoManager,
	}
}

// Login authenticates a user and returns access and refresh tokens
//
//nolint:nonamedreturns // we need named here
func (s *UsersService) Login(
	ctx context.Context,
	username, password string,
) (accessToken, refreshToken, sessionID string, isTmpPasswd bool, err error) {
	// Authenticate using the authentication provider chain
	user, err := s.authProvider.Authenticate(ctx, username, password)
	if err != nil {
		return "", "", "", false, fmt.Errorf("authentication failed: %w", err)
	}

	// Check if the user is active (should be checked by the provider, but just in case)
	if !user.IsActive {
		return "", "", "", false, domain.ErrInactiveUser
	}

	if user.TwoFAEnabled {
		sessionID = generate2FASession(user.ID, user.Username, time.Minute)

		return "", "", sessionID, false, domain.ErrTwoFARequired
	}

	// Generate tokens
	accessToken, err = s.tokenizer.AccessToken(user)
	if err != nil {
		return "", "", "", false, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.tokenizer.RefreshToken(user)
	if err != nil {
		return "", "", "", false, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.usersRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", "", "", false, fmt.Errorf("update last login at: %w", err)
	}

	return accessToken, refreshToken, "", user.IsTmpPassword, nil
}

// LoginReissue reissues a new access token using a valid refresh token
//
//nolint:nonamedreturns // we need named here
func (s *UsersService) LoginReissue(
	ctx context.Context,
	currRefreshToken string,
) (accessToken, refreshToken string, err error) {
	claims, err := s.tokenizer.VerifyToken(currRefreshToken, domain.TokenTypeRefresh)
	if err != nil {
		return "", "", fmt.Errorf("verify refresh token: %w", err)
	}

	user, err := s.usersRepo.GetByID(ctx, domain.UserID(claims.UserID))
	if err != nil {
		return "", "", fmt.Errorf("get user by uuid: %w", err)
	}

	if !user.IsActive {
		return "", "", domain.ErrInactiveUser
	}

	accessToken, err = s.tokenizer.AccessToken(&user)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.tokenizer.RefreshToken(&user)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.usersRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", "", fmt.Errorf("update last login at: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *UsersService) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	return s.usersRepo.GetByID(ctx, id)
}

func (s *UsersService) List(ctx context.Context) ([]domain.User, error) {
	// Permission check is performed in the handler layer
	// This allows for more flexible permission checks (e.g., membership.manage permission)
	return s.usersRepo.List(ctx)
}

// Create creates a new user. Only superusers can create new users.
func (s *UsersService) Create(
	ctx context.Context,
	currentUser domain.User,
	username, email, password string,
	isSuperuser bool,
) (domain.User, error) {
	// Check if the current user is a superuser
	if !currentUser.IsSuperuser {
		return domain.User{}, domain.ErrPermissionDenied
	}

	_, err := s.usersRepo.GetByUsername(ctx, username)
	if err == nil {
		return domain.User{}, domain.ErrUsernameAlreadyInUse
	}

	_, err = s.usersRepo.GetByEmail(ctx, email)
	if err == nil {
		return domain.User{}, domain.ErrEmailAlreadyInUse
	}

	passwordHash, err := passworder.PasswordHash(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	userDTO := domain.UserDTO{
		Username:      username,
		Email:         email,
		PasswordHash:  passwordHash,
		IsSuperuser:   isSuperuser,
		IsTmpPassword: true,
		IsExternal:    false,
	}

	user, err := s.usersRepo.Create(ctx, userDTO)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// SetSuperuserStatus sets or unsets the superuser status of a user.
// Only superusers can change the superuser status of other users.
// The admin user (username="admin") cannot have their superuser status modified.
func (s *UsersService) SetSuperuserStatus(
	ctx context.Context,
	id domain.UserID,
	isSuperuser bool,
) (domain.User, error) {
	// Get the current user from context
	currentUserID := appcontext.UserID(ctx)

	currentUser, err := s.usersRepo.GetByID(ctx, currentUserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get current user by id: %w", err)
	}

	// Check if the current user is a superuser
	if !currentUser.IsSuperuser {
		return domain.User{}, domain.ErrPermissionDenied
	}

	// Get the user to modify
	user, err := s.usersRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	if user.Username == "admin" {
		return domain.User{}, domain.ErrPermissionDenied
	}

	user.IsSuperuser = isSuperuser
	user.UpdatedAt = time.Now()

	// Save the changes
	if err := s.usersRepo.Update(ctx, &user); err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

// SetActiveStatus sets or unsets the active status of a user.
// Only superusers can change the active status of users.
func (s *UsersService) SetActiveStatus(ctx context.Context, id domain.UserID, isActive bool) (domain.User, error) {
	// Get the current user from context
	currentUserID := appcontext.UserID(ctx)

	currentUser, err := s.usersRepo.GetByID(ctx, currentUserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get current user by id: %w", err)
	}

	// Check if the current user is a superuser
	if !currentUser.IsSuperuser {
		return domain.User{}, domain.ErrPermissionDenied
	}

	// Get the user to modify
	user, err := s.usersRepo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	user.IsActive = isActive
	user.UpdatedAt = time.Now()

	if err := s.usersRepo.Update(ctx, &user); err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

// Delete deletes a user.
// Only superusers can delete users, and superusers cannot be deleted.
func (s *UsersService) Delete(ctx context.Context, id domain.UserID) error {
	// Get the current user from context
	currentUserID := appcontext.UserID(ctx)

	currentUser, err := s.usersRepo.GetByID(ctx, currentUserID)
	if err != nil {
		return fmt.Errorf("get current user by id: %w", err)
	}

	// Check if the current user is a superuser
	if !currentUser.IsSuperuser {
		return domain.ErrPermissionDenied
	}

	user, err := s.usersRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if user.IsSuperuser {
		return domain.ErrPermissionDenied
	}

	if err := s.usersRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

func (s *UsersService) UpdatePassword(ctx context.Context, id domain.UserID, oldPassword, newPassword string) error {
	user, err := s.usersRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if user.IsExternal {
		return domain.ErrPermissionDenied
	}

	isValid, err := passworder.ValidatePassword(oldPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("validate password: %w", err)
	}

	if !isValid {
		return domain.ErrInvalidPassword
	}

	passwordHash, err := passworder.PasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.usersRepo.UpdatePassword(ctx, id, passwordHash)
}

func (s *UsersService) ChangeTemporaryPassword(ctx context.Context, id domain.UserID, newPassword string) error {
	user, err := s.usersRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if user.IsExternal {
		return domain.ErrPermissionDenied
	}

	if !user.IsTmpPassword {
		return domain.ErrPermissionDenied
	}

	passwordHash, err := passworder.PasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.usersRepo.UpdatePassword(ctx, id, passwordHash)
}

func (s *UsersService) ForgotPassword(ctx context.Context, email string) error {
	slog.Debug("processing forgot password request")

	user, err := s.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			slog.Info("forgot password requested for non-existent email")
			// Don't reveal that the email doesn't exist
			return nil
		}

		slog.Error("failed to get user by email", "error", err)

		return fmt.Errorf("get user by email: %w", err)
	}

	slog.Debug("user found for password reset",
		"user_id", user.ID, "is_active", user.IsActive)

	if user.IsExternal {
		return domain.ErrPermissionDenied
	}

	if !user.IsActive {
		slog.Warn("inactive user tries to reset password", "user_id", user.ID)
	}

	token, ttl, err := s.tokenizer.ResetPasswordToken(&user)
	if err != nil {
		slog.Error("failed to generate reset password token",
			"user_id", user.ID, "error", err)

		return fmt.Errorf("generate reset password token: %w", err)
	}

	slog.Debug("reset password token generated", "user_id", user.ID, "token_ttl", ttl)

	err = s.emailer.SendResetPasswordEmail(ctx, email, token)
	if err != nil {
		slog.Error("failed to send reset password email", "user_id", user.ID, "error", err)

		return fmt.Errorf("send email: %w", err)
	}

	slog.Info("forgot password request processed successfully", "user_id", user.ID)

	return nil
}

func (s *UsersService) ResetPassword(ctx context.Context, token, newPassword string) error {
	claims, err := s.tokenizer.VerifyToken(token, domain.TokenTypeResetPassword)
	if err != nil {
		return fmt.Errorf("verify reset password token: %w", err)
	}

	user, err := s.usersRepo.GetByID(ctx, domain.UserID(claims.UserID))
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	passwordHash, err := passworder.PasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.usersRepo.UpdatePassword(ctx, user.ID, passwordHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

// UpdateLicenseAcceptance updates the license acceptance status for a user.
func (s *UsersService) UpdateLicenseAcceptance(ctx context.Context, userID domain.UserID, accepted bool) error {
	// Get the user
	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update the license acceptance status
	user.LicenseAccepted = accepted

	// Save the updated user
	err = s.usersRepo.Update(ctx, &user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// VerifyPassword verifies that the provided password is correct for the given user.
func (s *UsersService) VerifyPassword(ctx context.Context, userID domain.UserID, password string) error {
	// Get the user
	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return domain.ErrInactiveUser
	}

	// Verify the password using the passworder
	isValid, err := passworder.ValidatePassword(password, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("password validation error: %w", err)
	}

	if !isValid {
		return domain.ErrInvalidCredentials
	}

	return nil
}
