package rest

import (
	"errors"
	"net/http"
	"strconv"
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
	"github.com/rom8726/floxy-manager/internal/domain"
)

type Router struct {
	router             *httprouter.Router
	floxyMux           http.Handler
	staticMux          http.Handler
	authHandler        *handlers.AuthHandler
	passwordHandler    *handlers.PasswordHandler
	twoFAHandler       *handlers.TwoFAHandler
	ssoHandler         *handlers.SSOHandler
	tenantsHandler     *handlers.TenantsHandler
	projectsHandler    *handlers.ProjectsHandler
	workflowsHandler   *handlers.WorkflowsHandler
	usersHandler       *handlers.UsersHandler
	permissionsService contract.PermissionsService
}

func NewRouter(
	pool *pgxpool.Pool,
	usersService contract.UsersUseCase,
	tenantsRepo contract.TenantsRepository,
	projectsRepo contract.ProjectsRepository,
	workflowsRepo contract.WorkflowsRepository,
	permissionsService contract.PermissionsService,
	rolesRepo contract.RolesRepository,
	membershipsRepo contract.MembershipsRepository,
	membershipsSrv contract.MembershipsUseCase,
	ldapUseCase contract.LDAPSyncUseCase,
	settingsUseCase contract.SettingsUseCase,
	frontendURL string,
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
	ssoHandler := handlers.NewSSOHandler(usersService, frontendURL)
	tenantsHandler := handlers.NewTenantsHandler(tenantsRepo)
	projectsHandler := handlers.NewProjectsHandler(projectsRepo, permissionsService, rolesRepo, membershipsRepo)
	workflowsHandler := handlers.NewWorkflowsHandler(workflowsRepo, permissionsService)
	usersHandler := handlers.NewUsersHandler(usersService, projectsRepo, permissionsService)
	membershipsHandler := handlers.NewMembershipsHandler(membershipsSrv, usersService, permissionsService)
	ldapHandler := handlers.NewLDAPHandler(ldapUseCase, settingsUseCase)

	router.POST("/api/v1/auth/login", wrapHandler(authHandler.Login))
	router.POST("/api/v1/auth/refresh", wrapHandler(authHandler.Refresh))
	router.POST("/api/v1/auth/forgot-password", wrapHandler(passwordHandler.ForgotPassword))
	router.POST("/api/v1/auth/reset-password", wrapHandler(passwordHandler.ResetPassword))
	router.POST("/api/v1/auth/change-password", wrapHandler(passwordHandler.ChangePassword))

	router.GET("/api/v1/auth/sso/providers", wrapHandler(ssoHandler.GetProviders))
	router.POST("/api/v1/auth/sso/initiate", wrapHandler(ssoHandler.Initiate))
	router.GET("/api/v1/auth/sso/callback", wrapHandler(ssoHandler.Callback))
	router.GET("/api/v1/auth/saml/metadata", wrapHandler(ssoHandler.GetMetadata))
	router.POST("/api/v1/auth/saml/acs", wrapHandler(ssoHandler.ACS))

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
	router.GET("/api/v1/users", wrapHandler(usersHandler.ListUsers))
	router.POST("/api/v1/users", wrapHandler(usersHandler.CreateUser))
	router.PUT("/api/v1/users/:id/status", wrapHandler(usersHandler.UpdateUserStatus))
	router.DELETE("/api/v1/users/:id", wrapHandler(usersHandler.DeleteUser))

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

	// Memberships endpoints
	router.GET("/api/v1/projects/:id/memberships", wrapHandler(membershipsHandler.ListProjectMemberships))
	router.POST("/api/v1/projects/:id/memberships", wrapHandler(membershipsHandler.CreateProjectMembership))
	router.DELETE("/api/v1/projects/:id/memberships/:mid", wrapHandler(membershipsHandler.DeleteProjectMembership))
	router.GET("/api/v1/roles", wrapHandler(membershipsHandler.ListRoles))

	// LDAP endpoints
	router.GET("/api/v1/ldap/config", wrapHandler(ldapHandler.GetLDAPConfig))
	router.POST("/api/v1/ldap/config", wrapHandler(ldapHandler.UpdateLDAPConfig))
	router.DELETE("/api/v1/ldap/config", wrapHandler(ldapHandler.DeleteLDAPConfig))
	router.POST("/api/v1/ldap/test-connection", wrapHandler(ldapHandler.TestLDAPConnection))
	router.POST("/api/v1/ldap/sync/users", wrapHandler(ldapHandler.SyncLDAPUsers))
	router.DELETE("/api/v1/ldap/sync/cancel", wrapHandler(ldapHandler.CancelLDAPSync))
	router.GET("/api/v1/ldap/sync/status", wrapHandler(ldapHandler.GetLDAPSyncStatus))
	router.GET("/api/v1/ldap/sync/progress", wrapHandler(ldapHandler.GetLDAPSyncProgress))
	router.GET("/api/v1/ldap/sync/logs", wrapHandler(ldapHandler.GetLDAPSyncLogs))
	router.GET("/api/v1/ldap/sync/logs/:id", wrapHandler(ldapHandler.GetLDAPSyncLogDetails))
	router.GET("/api/v1/ldap/statistics", wrapHandler(ldapHandler.GetLDAPStatistics))

	floxyMux := floxyServer.Mux()

	staticMux := http.NewServeMux()
	staticFS := http.FileServer(http.Dir("./web/dist/"))
	staticMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/dist/assets/"))))
	staticMux.Handle("/bundle.js", staticFS)
	staticMux.Handle("/bundle.js.LICENSE.txt", staticFS)

	return &Router{
		router:             router,
		floxyMux:           floxyMux,
		staticMux:          staticMux,
		authHandler:        authHandler,
		passwordHandler:    passwordHandler,
		twoFAHandler:       twoFAHandler,
		ssoHandler:         ssoHandler,
		tenantsHandler:     tenantsHandler,
		projectsHandler:    projectsHandler,
		workflowsHandler:   workflowsHandler,
		usersHandler:       usersHandler,
		permissionsService: permissionsService,
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
		// Enforce strict RBAC for mutating plugin API calls
		if isMutatingMethod(req.Method) {
			// Require auth: user must be set in context by outer middleware
			if appcontext.UserID(req.Context()) == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			projID, ok := extractProjectID(req)
			if !ok {
				http.Error(w, "project_id is required (query ?project_id=... or header X-Project-ID or Referer path)", http.StatusBadRequest)
				return
			}

			if err := r.permissionsService.CanManageProject(req.Context(), domain.ProjectID(projID)); err != nil {
				if errors.Is(err, domain.ErrPermissionDenied) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				if errors.Is(err, domain.ErrUserNotFound) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		r.floxyMux.ServeHTTP(w, req)
		return
	}

	http.ServeFile(w, req, "./web/dist/index.html")
}

func isMutatingMethod(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// extractProjectID tries multiple sources to determine project ID for RBAC checks:
// 1) Query param "project_id"
// 2) Header "X-Project-ID"
// 3) Referer path segment: /tenants/{tid}/projects/{pid}/...
func extractProjectID(r *http.Request) (int, bool) {
	// Query param
	if v := r.URL.Query().Get("project_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil && id > 0 {
			return id, true
		}
	}
	// Header
	if v := r.Header.Get("X-Project-ID"); v != "" {
		if id, err := strconv.Atoi(v); err == nil && id > 0 {
			return id, true
		}
	}
	// Referer parsing
	if ref := r.Referer(); ref != "" {
		if pid, ok := parseProjectIDFromURLPath(ref); ok {
			return pid, true
		}
	}
	return 0, false
}

func parseProjectIDFromURLPath(rawURL string) (int, bool) {
	// We only need the path component; be tolerant to raw paths too
	u, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err == nil && u.URL != nil {
		return findProjectIDInPath(u.URL.Path)
	}
	// Fallback: treat rawURL as path
	return findProjectIDInPath(rawURL)
}

func findProjectIDInPath(path string) (int, bool) {
	segs := strings.Split(path, "/")
	for i := 0; i < len(segs); i++ {
		if segs[i] == "projects" && i+1 < len(segs) {
			if id, err := strconv.Atoi(segs[i+1]); err == nil && id > 0 {
				return id, true
			}
		}
	}
	return 0, false
}

func wrapHandler(fn http.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := appcontext.WithParams(r.Context(), ps)
		fn(w, r.WithContext(ctx))
	}
}
