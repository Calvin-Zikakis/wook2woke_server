package server

import (
	"bytes"
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()

	tmpDir := t.TempDir()
	photoDir := filepath.Join(tmpDir, "photos")
	os.MkdirAll(photoDir, 0755)

	db, err := sql.Open("sqlite3", filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
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
	`)
	if err != nil {
		t.Fatal(err)
	}

	return &Server{
		db: db,
		cfg: Config{
			APIKey:        "test-api-key",
			ViewPassword:  "test-password",
			AdminPassword: "test-admin-password",
			PhotoDir:      photoDir,
			SessionTTL:    time.Hour,
		},
		sessions: make(map[string]sessionData),
		mu:       sync.RWMutex{},
	}
}

func writeTestPhoto(t *testing.T, dir, name string) {
	t.Helper()
	// Minimal valid JPEG bytes
	err := os.WriteFile(filepath.Join(dir, name), []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0xFF, 0xD9}, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func newUploadRequest(t *testing.T, apiKey string, wokeScore int, description string, photoFilename string) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	w.WriteField("wokeScore", fmt.Sprintf("%d", wokeScore))
	w.WriteField("description", description)

	part, err := w.CreateFormFile("photo", photoFilename)
	if err != nil {
		t.Fatal(err)
	}
	// Minimal valid JPEG bytes
	part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0xFF, 0xD9})
	w.Close()

	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	return req
}
