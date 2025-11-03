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
	"github.com/rom8726/floxy-manager/internal/contract"
)

type Router struct {
	router          *httprouter.Router
	floxyMux        http.Handler
	staticMux       http.Handler
	authHandler     *handlers.AuthHandler
	passwordHandler *handlers.PasswordHandler
	twoFAHandler    *handlers.TwoFAHandler
	ssoHandler      *handlers.SSOHandler
}

func NewRouter(
	pool *pgxpool.Pool,
	usersService contract.UsersUseCase,
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

	router.POST("/api/v1/auth/login", wrapHandler(authHandler.Login))
	router.POST("/api/v1/auth/refresh", wrapHandler(authHandler.Refresh))
	router.POST("/api/v1/auth/forgot-password", wrapHandler(passwordHandler.ForgotPassword))
	router.POST("/api/v1/auth/reset-password", wrapHandler(passwordHandler.ResetPassword))

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

	floxyMux := floxyServer.Mux()

	staticMux := http.NewServeMux()
	staticFS := http.FileServer(http.Dir("./web/dist/"))
	staticMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/dist/assets/"))))
	staticMux.Handle("/bundle.js", staticFS)
	staticMux.Handle("/bundle.js.LICENSE.txt", staticFS)

	return &Router{
		router:          router,
		floxyMux:        floxyMux,
		staticMux:       staticMux,
		authHandler:     authHandler,
		passwordHandler: passwordHandler,
		twoFAHandler:    twoFAHandler,
		ssoHandler:      ssoHandler,
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
		fn(w, r)
	}
}
