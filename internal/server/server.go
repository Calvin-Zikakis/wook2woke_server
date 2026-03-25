package server

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"
)

type Entry struct {
	ID          int
	WokeScore   int
	Description string
	PhotoPath   string
	CreatedAt   string
}

type Config struct {
	Port            string
	APIKey          string
	ViewPassword    string
	PhotoDir        string
	DBPath          string
	SessionTTL      time.Duration
	AnthropicAPIKey string
}

type Server struct {
	db       *sql.DB
	cfg      Config
	sessions map[string]time.Time
	mu       sync.RWMutex
	// claudeFn can be set in tests to avoid real API calls
	claudeFn func(ctx context.Context, imgData []byte) (*claudeResult, error)
}

func New(cfg Config, db *sql.DB) *Server {
	return &Server{
		db:       db,
		cfg:      cfg,
		sessions: make(map[string]time.Time),
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/upload", s.handleUpload)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("GET /", s.requireAuth(s.handleGallery))
	mux.HandleFunc("GET /photos/", s.requireAuth(s.handlePhoto))
	mux.HandleFunc("GET /api/entries", s.requireAuth(s.handleAPIEntries))
	mux.HandleFunc("POST /api/entries/{id}/rescore", s.requireAuth(s.handleRescore))
	mux.HandleFunc("GET /api/entries/{id}/rescores", s.requireAuth(s.handleGetRescores))
}
