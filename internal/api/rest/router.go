package rest

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/rom8726/floxy"
	"github.com/rom8726/floxy/api"
	"github.com/rom8726/floxy/plugins/api/abort"
	"github.com/rom8726/floxy/plugins/api/cancel"
	"github.com/rom8726/floxy/plugins/api/cleanup"
	"github.com/rom8726/floxy/plugins/api/dlq"
	human_decision "github.com/rom8726/floxy/plugins/api/human-decision"

	"github.com/rom8726/floxy-manager/internal/api/rest/handlers"
	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/repository/workflows"
)

type Router struct {
	router           *httprouter.Router
	floxyMux         http.Handler
	staticMux        http.Handler
	authHandler      *handlers.AuthHandler
	passwordHandler  *handlers.PasswordHandler
	twoFAHandler     *handlers.TwoFAHandler
	ssoHandler       *handlers.SSOHandler
	tenantsHandler   *handlers.TenantsHandler
	projectsHandler  *handlers.ProjectsHandler
	workflowsHandler *handlers.WorkflowsHandler
	usersHandler     *handlers.UsersHandler
}

func NewRouter(
	pool *pgxpool.Pool,
	usersService contract.UsersUseCase,
	tenantsRepo contract.TenantsRepository,
	projectsRepo contract.ProjectsRepository,
	workflowsRepo *workflows.Repository,
	permissionsService contract.PermissionsService,
	rolesRepo contract.RolesRepository,
	membershipsRepo contract.MembershipsRepository,
) (*Router, error) {
	store := floxy.NewStore(pool)
	engine := floxy.NewEngine(pool)

	humanDecisionPlugin := human_decision.New(engine, store, func(*http.Request) (string, error) {
		return "admin", nil
	})
	cancelPlugin := cancel.New(engine, func(req *http.Request) (string, error) {
		return "admin", nil
	})
	abortPlugin := abort.New(engine, func(req *http.Request) (string, error) {
		return "admin", nil
	})
	dlqPlugin := dlq.New(engine, store)
	cleanupPlugin := cleanup.New(store)

	floxyServer := api.New(engine, store, api.WithPlugins(
		humanDecisionPlugin,
		cancelPlugin,
		abortPlugin,
		dlqPlugin,
		cleanupPlugin,
	))

	router := httprouter.New()

	authHandler := handlers.NewAuthHandler(usersService)
	passwordHandler := handlers.NewPasswordHandler(usersService)
	twoFAHandler := handlers.NewTwoFAHandler(usersService)
	ssoHandler := handlers.NewSSOHandler(usersService)
	tenantsHandler := handlers.NewTenantsHandler(tenantsRepo)
	projectsHandler := handlers.NewProjectsHandler(projectsRepo, permissionsService, rolesRepo, membershipsRepo)
	workflowsHandler := handlers.NewWorkflowsHandler(workflowsRepo)
	usersHandler := handlers.NewUsersHandler(usersService, projectsRepo, permissionsService)

	router.POST("/api/v1/auth/login", wrapHandler(authHandler.Login))
	router.POST("/api/v1/auth/refresh", wrapHandler(authHandler.Refresh))
	router.POST("/api/v1/auth/forgot-password", wrapHandler(passwordHandler.ForgotPassword))
	router.POST("/api/v1/auth/reset-password", wrapHandler(passwordHandler.ResetPassword))
	router.POST("/api/v1/auth/change-password", wrapHandler(passwordHandler.ChangePassword))

	router.GET("/api/v1/auth/sso/providers", wrapHandler(ssoHandler.GetProviders))
	router.POST("/api/v1/auth/sso/initiate", wrapHandler(ssoHandler.Initiate))
	router.GET("/api/v1/auth/sso/callback", wrapHandler(ssoHandler.Callback))
	router.GET("/api/v1/auth/sso/metadata", wrapHandler(ssoHandler.GetMetadata))

	router.POST("/api/v1/auth/2fa/verify", wrapHandler(twoFAHandler.Verify2FA))
	router.POST("/api/v1/auth/2fa/setup", wrapHandler(twoFAHandler.Setup2FA))
	router.POST("/api/v1/auth/2fa/confirm", wrapHandler(twoFAHandler.Confirm2FA))
	router.POST("/api/v1/auth/2fa/send-code", wrapHandler(twoFAHandler.Send2FACode))
	router.POST("/api/v1/auth/2fa/disable", wrapHandler(twoFAHandler.Disable2FA))
	router.POST("/api/v1/auth/2fa/reset", wrapHandler(twoFAHandler.Reset2FA))

	router.GET("/api/v1/tenants", wrapHandler(tenantsHandler.List))
	router.POST("/api/v1/tenants", wrapHandler(tenantsHandler.Create))
	router.PUT("/api/v1/tenants/:id", wrapHandler(tenantsHandler.Update))
	router.DELETE("/api/v1/tenants/:id", wrapHandler(tenantsHandler.Delete))
	router.GET("/api/v1/projects", wrapHandler(projectsHandler.List))
	router.POST("/api/v1/projects", wrapHandler(projectsHandler.Create))
	router.PUT("/api/v1/projects/:id", wrapHandler(projectsHandler.Update))
	router.DELETE("/api/v1/projects/:id", wrapHandler(projectsHandler.Delete))

	// User account endpoints
	router.GET("/api/v1/users/me", wrapHandler(usersHandler.GetCurrentUser))
	router.GET("/api/v1/users/me/projects", wrapHandler(usersHandler.GetMyProjects))
	router.POST("/api/v1/users/me/password", wrapHandler(usersHandler.UpdatePassword))

	// Workflows endpoints
	router.GET("/api/v1/workflows", wrapHandler(workflowsHandler.ListWorkflows))
	router.GET("/api/v1/workflows/:id", wrapHandler(workflowsHandler.GetWorkflow))
	router.GET("/api/v1/workflows/:id/instances", wrapHandler(workflowsHandler.ListWorkflowInstances))
	router.GET("/api/v1/instances", wrapHandler(workflowsHandler.ListInstances))
	router.GET("/api/v1/instances/:id", wrapHandler(workflowsHandler.GetInstance))
	router.GET("/api/v1/active-workflows", wrapHandler(workflowsHandler.ListActiveWorkflows))
	router.GET("/api/v1/instances/:id/steps", wrapHandler(workflowsHandler.ListInstanceSteps))
	router.GET("/api/v1/instances/:id/events", wrapHandler(workflowsHandler.ListInstanceEvents))
	router.GET("/api/v1/stats", wrapHandler(workflowsHandler.ListStats))
	router.GET("/api/v1/dlq", wrapHandler(workflowsHandler.ListDLQ))
	router.GET("/api/v1/dlq/:id", wrapHandler(workflowsHandler.GetDLQItem))

	floxyMux := floxyServer.Mux()

	staticMux := http.NewServeMux()
	staticFS := http.FileServer(http.Dir("./web/dist/"))
	staticMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/dist/assets/"))))
	staticMux.Handle("/bundle.js", staticFS)
	staticMux.Handle("/bundle.js.LICENSE.txt", staticFS)

	return &Router{
		router:           router,
		floxyMux:         floxyMux,
		staticMux:        staticMux,
		authHandler:      authHandler,
		passwordHandler:  passwordHandler,
		twoFAHandler:     twoFAHandler,
		ssoHandler:       ssoHandler,
		tenantsHandler:   tenantsHandler,
		projectsHandler:  projectsHandler,
		workflowsHandler: workflowsHandler,
		usersHandler:     usersHandler,
	}, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if strings.HasPrefix(path, "/api/v1/") {
		r.router.ServeHTTP(w, req)
		return
	}

	if path == "/assets/" || strings.HasPrefix(path, "/assets/") ||
		path == "/bundle.js" || path == "/bundle.js.LICENSE.txt" {
		r.staticMux.ServeHTTP(w, req)
		return
	}

	if strings.HasPrefix(path, "/api/") {
		r.floxyMux.ServeHTTP(w, req)
		return
	}

	http.ServeFile(w, req, "./web/dist/index.html")
}

func wrapHandler(fn http.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := appcontext.WithParams(r.Context(), ps)
		fn(w, r.WithContext(ctx))
	}
}
