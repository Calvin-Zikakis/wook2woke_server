package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"wook2woke_server/internal/server"
)

func main() {
	cfg := server.Config{
		Port:            envOrDefault("PORT", "8080"),
		APIKey:          requireEnv("API_KEY"),
		ViewPassword:    os.Getenv("VIEW_PASSWORD"),
		AdminPassword:   requireEnv("ADMIN_PASSWORD"),
		PhotoDir:        envOrDefault("PHOTO_DIR", "./photos"),
		DBPath:          envOrDefault("DB_PATH", "./wook2woke.db"),
		SessionTTL:      24 * time.Hour,
		AnthropicAPIKey: requireEnv("ANTHROPIC_API_KEY"),
	}

	if err := os.MkdirAll(cfg.PhotoDir, 0755); err != nil {
		log.Fatalf("failed to create photo dir: %v", err)
	}

	db, err := sql.Open("sqlite3", cfg.DBPath+"?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			woke_score INTEGER NOT NULL,
			description TEXT NOT NULL,
			photo_path TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS rescores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entry_id INTEGER NOT NULL REFERENCES entries(id),
			woke_score INTEGER NOT NULL,
			subject TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	srv := server.New(cfg, db)

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	log.Printf("wook2woke server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
