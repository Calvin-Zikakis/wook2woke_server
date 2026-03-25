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
	VoteAvg     float64
	VoteCount   int
	UserVote    int // 0 = not voted
}

type Config struct {
	Port            string
	APIKey          string
	ViewPassword    string
	AdminPassword   string
	PhotoDir        string
	DBPath          string
	SessionTTL      time.Duration
	AnthropicAPIKey string
}

type sessionData struct {
	expiry time.Time
	role   string // "viewer" or "admin"
}

type Server struct {
	db       *sql.DB
	cfg      Config
	sessions map[string]sessionData
	mu       sync.RWMutex
	// claudeFn can be set in tests to avoid real API calls
	claudeFn func(ctx context.Context, imgData []byte) (*claudeResult, error)
}

func New(cfg Config, db *sql.DB) *Server {
	return &Server{
		db:       db,
		cfg:      cfg,
		sessions: make(map[string]sessionData),
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/upload", s.handleUpload)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /logout", s.handleLogout)
	mux.HandleFunc("GET /", s.requireAuth(s.handleGallery))
	mux.HandleFunc("GET /photos/", s.requireAuth(s.handlePhoto))
	mux.HandleFunc("GET /api/entries", s.requireAuth(s.handleAPIEntries))
	mux.HandleFunc("GET /live", s.requireAuth(s.handleLive))
	mux.HandleFunc("POST /api/entries/{id}/rescore", s.requireAdmin(s.handleRescore))
	mux.HandleFunc("GET /api/entries/{id}/rescores", s.requireAuth(s.handleGetRescores))
	mux.HandleFunc("DELETE /api/entries/{id}", s.requireAdmin(s.handleDelete))
	mux.HandleFunc("PUT /api/entries/{id}", s.requireAdmin(s.handlePromoteRescore))
	mux.HandleFunc("POST /api/entries/{id}/vote", s.requireAuth(s.handleVote))
}
