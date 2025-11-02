package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rom8726/floxy"
	"github.com/rom8726/floxy/api"
	"github.com/rom8726/floxy/plugins/api/abort"
	"github.com/rom8726/floxy/plugins/api/cancel"
	"github.com/rom8726/floxy/plugins/api/cleanup"
	"github.com/rom8726/floxy/plugins/api/dlq"
	human_decision "github.com/rom8726/floxy/plugins/api/human-decision"

	"github.com/rom8726/floxy-manager/internal/config"
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
	dlqPlugin := dlq.New(engine, store)
	cleanupPlugin := cleanup.New(store)

	floxyServer := api.New(engine, store, api.WithPlugins(
		humanDecisionPlugin,
		cancelPlugin,
		abortPlugin,
		dlqPlugin,
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

	// Handle static assets - register BEFORE catch-all route
	// This must be registered first to avoid being caught by the SPA fallback
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/dist/assets/"))))
	mux.Handle("/bundle.js", staticFS)
	mux.Handle("/bundle.js.LICENSE.txt", staticFS)

	// Auth API - Login endpoint (stub/mock)
	mux.HandleFunc("/api/auth/login", s.handleLogin)

	// Serve index.html for all other routes (SPA fallback)
	// This catch-all must be registered LAST
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Skip static file requests - these should be handled by registered routes above
		// Just serve index.html for SPA routing
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

// handleLogin is a stub/mock login endpoint
// TODO: Replace with actual authentication logic
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request
	var loginReq struct {
		UsernameOrEmail string `json:"username_or_email"`
		Password        string `json:"password"`
	}

	if err := json.Unmarshal(body, &loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if loginReq.UsernameOrEmail == "" || loginReq.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Username/email and password are required",
		})
		return
	}

	// Mock authentication - accept any credentials for now
	// TODO: Implement real authentication
	// For now, this is just a stub that always succeeds if fields are provided

	// Extract username from email if needed
	username := loginReq.UsernameOrEmail
	if strings.Contains(loginReq.UsernameOrEmail, "@") {
		parts := strings.Split(loginReq.UsernameOrEmail, "@")
		username = parts[0]
	}

	// Generate mock response
	response := map[string]interface{}{
		"token": "mock-auth-token-" + username,
		"user": map[string]interface{}{
			"username": username,
			"email": func() string {
				if strings.Contains(loginReq.UsernameOrEmail, "@") {
					return loginReq.UsernameOrEmail
				}
				return ""
			}(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
