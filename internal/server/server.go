package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rom8726/floxy"
	"github.com/rom8726/floxy/api"
	"github.com/rom8726/floxy/plugins/api/abort"
	"github.com/rom8726/floxy/plugins/api/cancel"
	"github.com/rom8726/floxy/plugins/api/cleanup"
	human_decision "github.com/rom8726/floxy/plugins/api/human-decision"

	"github.com/rom8726/floxy-ui/internal/config"
)

type Server struct {
	config *config.Config
	pool   *pgxpool.Pool
	engine *floxy.Engine
	floxy  *api.Server
}

func New(cfg *config.Config) (*Server, error) {
	// Build database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	// Connect to database
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create a floxy server
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
	cleanupPlugin := cleanup.New(store)

	floxyServer := api.New(engine, store, api.WithPlugins(
		humanDecisionPlugin,
		cancelPlugin,
		abortPlugin,
		cleanupPlugin,
	))

	return &Server{
		config: cfg,
		pool:   pool,
		engine: engine,
		floxy:  floxyServer,
	}, nil
}

func (s *Server) Start() error {
	mux := s.floxy.Mux()

	// Serve static files from web/dist directory
	staticFS := http.FileServer(http.Dir("./web/dist/"))

	// Handle static assets (JS, CSS, etc.)
	mux.Handle("/bundle.js", staticFS)
	mux.Handle("/bundle.js.LICENSE.txt", staticFS)

	// Serve index.html for all other routes (SPA fallback)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for a static file
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, "./web/dist/index.html")
			return
		}

		// For all other routes, serve index.html (SPA fallback)
		http.ServeFile(w, r, "./web/dist/index.html")
	})

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	return srv.ListenAndServe()
}

func (s *Server) Close() {
	if s.engine != nil {
		s.engine.Shutdown()
	}
	if s.pool != nil {
		s.pool.Close()
	}
}
