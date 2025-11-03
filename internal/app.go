package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"path"
	"time"

	"github.com/rom8726/floxy-manager/internal/api/rest"
	"github.com/rom8726/floxy-manager/internal/api/rest/middlewares"
	"github.com/rom8726/floxy-manager/internal/config"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/internal/repository/ldapsynclogs"
	"github.com/rom8726/floxy-manager/internal/repository/ldapsyncstats"
	"github.com/rom8726/floxy-manager/internal/repository/licenses"
	"github.com/rom8726/floxy-manager/internal/repository/productinfo"
	"github.com/rom8726/floxy-manager/internal/repository/projects"
	"github.com/rom8726/floxy-manager/internal/repository/rbac"
	"github.com/rom8726/floxy-manager/internal/repository/settings"
	"github.com/rom8726/floxy-manager/internal/repository/tenants"
	"github.com/rom8726/floxy-manager/internal/repository/users"
	"github.com/rom8726/floxy-manager/internal/repository/workflows"
	ratelimiter2fa "github.com/rom8726/floxy-manager/internal/services/2fa/ratelimiter"
	"github.com/rom8726/floxy-manager/internal/services/email"
	"github.com/rom8726/floxy-manager/internal/services/ldap"
	"github.com/rom8726/floxy-manager/internal/services/permissions"
	ssoprovidermanager "github.com/rom8726/floxy-manager/internal/services/sso/provider-manager"
	samlprovider "github.com/rom8726/floxy-manager/internal/services/sso/saml"
	"github.com/rom8726/floxy-manager/internal/services/tokenizer"
	ldapusecase "github.com/rom8726/floxy-manager/internal/usecases/ldap"
	projectsusecase "github.com/rom8726/floxy-manager/internal/usecases/projects"
	rbacusecase "github.com/rom8726/floxy-manager/internal/usecases/rbac"
	settingsusecase "github.com/rom8726/floxy-manager/internal/usecases/settings"
	usersusecase "github.com/rom8726/floxy-manager/internal/usecases/users"
	"github.com/rom8726/floxy-manager/pkg/db"
	"github.com/rom8726/floxy-manager/pkg/httpserver"
	pkgmiddlewares "github.com/rom8726/floxy-manager/pkg/httpserver/middlewares"
	"github.com/rom8726/floxy-manager/pkg/passworder"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rom8726/di"
	"golang.org/x/sync/errgroup"
)

const (
	ctxTimeout = 10 * time.Second
)

type App struct {
	Config *config.Config
	Logger *slog.Logger

	PostgresPool *pgxpool.Pool

	APIServer *httpserver.Server

	container *di.Container
	diApp     *di.App
}

func NewApp(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()

	pgPool, err := newPostgresConnPool(ctx, &cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	container := di.New()
	diApp := di.NewApp(container)

	app := &App{
		Config:       cfg,
		Logger:       logger,
		container:    container,
		diApp:        diApp,
		PostgresPool: pgPool,
	}

	app.registerComponents()

	app.APIServer, err = app.newAPIServer()
	if err != nil {
		return nil, fmt.Errorf("create API server: %w", err)
	}

	return app, nil
}

func (app *App) RegisterComponent(constructor any) *di.Provider {
	return app.container.Provide(constructor)
}

func (app *App) ResolveComponent(target any) error {
	return app.container.Resolve(target)
}

func (app *App) ResolveComponentsToStruct(target any) error {
	return app.container.ResolveToStruct(target)
}

func (app *App) Run(ctx context.Context) error {
	// Check and create superuser if needed
	if app.Config.AdminEmail != "" {
		if err := app.ensureSuperuser(ctx); err != nil {
			app.Logger.Error("Failed to ensure superuser exists", "error", err)
		}
	}

	techServer, err := app.newTechServer()
	if err != nil {
		return fmt.Errorf("create tech server: %w", err)
	}

	app.Logger.Info("Start API server")

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return app.APIServer.ListenAndServe(groupCtx) })
	group.Go(func() error { return techServer.ListenAndServe(groupCtx) })
	group.Go(func() error { return app.diApp.Run(groupCtx) })

	return group.Wait()
}

func (app *App) Close() {
	if app.PostgresPool != nil {
		app.PostgresPool.Close()
	}
}

func (app *App) registerComponent(constructor any) *di.Provider {
	return app.container.Provide(constructor)
}

func (app *App) registerComponents() {
	app.registerComponent(db.NewTxManager).Arg(app.PostgresPool)

	// Register repositories
	app.registerComponent(projects.New).Arg(app.PostgresPool)
	app.registerComponent(users.New).Arg(app.PostgresPool)
	app.registerComponent(tenants.New).Arg(app.PostgresPool)
	app.registerComponent(ldapsyncstats.New).Arg(app.PostgresPool)
	app.registerComponent(ldapsynclogs.New).Arg(app.PostgresPool)
	app.registerComponent(licenses.New).Arg(app.PostgresPool)
	app.registerComponent(productinfo.New).Arg(app.PostgresPool)
	app.registerComponent(settings.New).Arg(app.PostgresPool)
	app.registerComponent(workflows.New).Arg(app.PostgresPool)
	// Register RBAC repositories
	app.registerComponent(rbac.NewRoles).Arg(app.PostgresPool)
	app.registerComponent(rbac.NewPermissions).Arg(app.PostgresPool)
	app.registerComponent(rbac.NewMemberships).Arg(app.PostgresPool)

	// Register permissions service
	app.registerComponent(permissions.New)

	app.registerComponent(email.New).Arg(&email.Config{
		SMTPHost:      app.Config.Mailer.Addr,
		Username:      app.Config.Mailer.User,
		Password:      app.Config.Mailer.Password,
		CertFile:      app.Config.Mailer.CertFile,
		KeyFile:       app.Config.Mailer.KeyFile,
		AllowInsecure: app.Config.Mailer.AllowInsecure,
		UseTLS:        app.Config.Mailer.UseTLS,
		BaseURL:       app.Config.FrontendURL,
		From:          app.Config.Mailer.From,
	})

	// Register use cases
	app.registerComponent(projectsusecase.New)
	app.registerComponent(ldapusecase.New)
	app.registerComponent(rbacusecase.New)
	app.registerComponent(settingsusecase.New).Arg(app.Config.SecretKey)

	// Register LDAP service
	app.registerComponent(ldap.New)

	var ldapService contract.LDAPService
	if err := app.container.Resolve(&ldapService); err != nil {
		panic(err)
	}

	// Initialize SSO provider manager
	app.registerComponent(ssoprovidermanager.New)

	// Initialize SAML provider
	app.registerComponent(samlprovider.New).Arg(&samlprovider.SAMLParams{
		Name:        domain.SSOProviderNameADSaml,
		DisplayName: "Sign in with Active Directory",
		IconURL:     "",
		Config: &domain.SAMLConfig{
			Enabled:          app.Config.SAML.Enabled,
			CreateCerts:      app.Config.SAML.CreateCerts,
			EntityID:         app.Config.SAML.EntityID,
			CertificatePath:  app.Config.SAML.CertificatePath,
			PrivateKeyPath:   app.Config.SAML.PrivateKeyPath,
			IDPMetadataURL:   app.Config.SAML.IDPMetadataURL,
			SSOURL:           app.Config.SAML.SSOURL,
			AttributeMapping: app.Config.SAML.AttributeMapping,
			CallbackURL:      path.Join(app.Config.FrontendURL, "/api/v1/auth/sso/callback"),
			PublicRootURL:    app.Config.FrontendURL,
			SkipTLSVerify:    app.Config.SAML.SkipTLSVerify,
		},
	})

	var samlProvider *samlprovider.SAMLProvider
	if err := app.container.Resolve(&samlProvider); err != nil {
		panic(err)
	}

	app.registerComponent(usersusecase.New).Arg([]usersusecase.AuthProvider{
		ldap.NewAuthService(ldapService.(*ldap.Service)), //nolint:forcetypeassert // ldapService guaranteed
	})

	// Register services
	app.registerComponent(tokenizer.New).Arg(&tokenizer.ServiceParams{
		SecretKey:        []byte(app.Config.JWTSecretKey),
		AccessTTL:        app.Config.AccessTokenTTL,
		RefreshTTL:       app.Config.RefreshTokenTTL,
		ResetPasswordTTL: app.Config.ResetPasswordTTL,
	})
	app.registerComponent(ratelimiter2fa.New)
}

func (app *App) newAPIServer() (*httpserver.Server, error) {
	cfg := app.Config.APIServer

	var tokenizerSrv contract.Tokenizer
	if err := app.container.Resolve(&tokenizerSrv); err != nil {
		return nil, fmt.Errorf("resolve tokenizer service component: %w", err)
	}

	var usersSrv contract.UsersUseCase
	if err := app.container.Resolve(&usersSrv); err != nil {
		return nil, fmt.Errorf("resolve users service component: %w", err)
	}

	app.registerComponent(rest.NewRouter).Arg(app.PostgresPool).Arg(app.Config.FrontendURL)
	var apiRouter *rest.Router
	if err := app.container.Resolve(&apiRouter); err != nil {
		return nil, fmt.Errorf("resolve api router component: %w", err)
	}

	handler := pkgmiddlewares.CORSMdw(
		middlewares.WithRawRequest(
			middlewares.RequestIDMdw(
				middlewares.AuthMiddleware(tokenizerSrv, usersSrv)(
					apiRouter,
				),
			),
		),
	)

	lis, err := net.Listen("tcp", cfg.Addr) //nolint:noctx // need to refactor
	if err != nil {
		return nil, fmt.Errorf("listen %q: %w", cfg.Addr, err)
	}

	return &httpserver.Server{
		Listener:     lis,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      handler,
	}, nil
}

func (app *App) newTechServer() (*httpserver.Server, error) {
	cfg := app.Config.TechServer
	lis, err := net.Listen("tcp", cfg.Addr) //nolint:noctx // need to refactor

	if err != nil {
		return nil, fmt.Errorf("listen %q: %w", cfg.Addr, err)
	}

	router := httprouter.New()
	router.Handle(http.MethodGet, "/health",
		func(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte("ok"))
		},
	)

	router.Handler(http.MethodGet, "/metrics", promhttp.Handler())

	router.HandlerFunc(http.MethodGet, "/debug/pprof", pprof.Index)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/cmdline", pprof.Cmdline)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/profile", pprof.Profile)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/symbol", pprof.Symbol)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/trace", pprof.Trace)
	router.Handler(http.MethodGet, "/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handler(http.MethodGet, "/debug/pprof/block", pprof.Handler("block"))
	router.Handler(http.MethodGet, "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handler(http.MethodGet, "/debug/pprof/heap", pprof.Handler("heap"))
	router.Handler(http.MethodGet, "/debug/pprof/mutex", pprof.Handler("mutex"))
	router.Handler(http.MethodGet, "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	return &httpserver.Server{
		Listener:     lis,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      router,
	}, nil
}

// ensureSuperuser checks if a user with the admin email exists and creates one if not.
func (app *App) ensureSuperuser(ctx context.Context) error {
	app.Logger.Info("Checking if superuser exists")

	var usersRepo contract.UsersRepository
	if err := app.container.Resolve(&usersRepo); err != nil {
		return fmt.Errorf("resolve users repository: %w", err)
	}

	// Check if user with admin email already exists
	_, err := usersRepo.GetByEmail(ctx, app.Config.AdminEmail)
	if err == nil {
		// User already exists
		app.Logger.Info("Superuser already exists")

		return nil
	}

	// Extract username from email (part before @)
	username := app.Config.AdminEmail
	for i, c := range username {
		if c == '@' {
			username = username[:i]

			break
		}
	}

	// Hash the temporary password
	passwordHash, err := passworder.PasswordHash(app.Config.AdminTmpPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Create the superuser
	userDTO := domain.UserDTO{
		Username:      username,
		Email:         app.Config.AdminEmail,
		PasswordHash:  passwordHash,
		IsSuperuser:   true,
		IsTmpPassword: true,
		IsExternal:    false,
	}

	user, err := usersRepo.Create(ctx, userDTO)
	if err != nil {
		return fmt.Errorf("create superuser: %w", err)
	}

	app.Logger.Info("Created superuser", "id", user.ID, "username", user.Username)

	return nil
}

func newPostgresConnPool(ctx context.Context, cfg *config.Postgres) (*pgxpool.Pool, error) {
	pgCfg, err := pgxpool.ParseConfig(cfg.ConnStringWithPoolSize())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pgCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	pgCfg.MaxConnLifetimeJitter = time.Second * 5
	pgCfg.MaxConnIdleTime = cfg.MaxIdleConnTime
	pgCfg.HealthCheckPeriod = time.Second * 5

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
