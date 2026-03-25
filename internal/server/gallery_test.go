package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGallery_Empty(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	srv.handleGallery(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "No entries yet") {
		t.Fatal("expected empty state message")
	}
}

func TestGallery_WithEntries(t *testing.T) {
	srv := newTestServer(t)

	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 99, "extremely woke", "test.jpg")
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 10, "not very woke", "test2.jpg")

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	srv.handleGallery(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "extremely woke") {
		t.Fatal("expected first entry description")
	}
	if !strings.Contains(body, "not very woke") {
		t.Fatal("expected second entry description")
	}
	if !strings.Contains(body, "99") {
		t.Fatal("expected score 99")
	}
}

func TestGallery_UnknownPath(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/does-not-exist", nil)
	rr := httptest.NewRecorder()

	srv.handleGallery(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown path, got %d", rr.Code)
	}
}

func TestAPIEntries_Empty(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/entries", nil)
	rr := httptest.NewRecorder()

	srv.handleAPIEntries(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var entries []Entry
	json.NewDecoder(rr.Body).Decode(&entries)
	if entries != nil {
		t.Fatalf("expected nil entries, got %d", len(entries))
	}
}

func TestAPIEntries_WithData(t *testing.T) {
	srv := newTestServer(t)
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 77, "mid woke", "mid.jpg")

	req := httptest.NewRequest("GET", "/api/entries", nil)
	rr := httptest.NewRecorder()

	srv.handleAPIEntries(rr, req)

	var entries []Entry
	json.NewDecoder(rr.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].WokeScore != 77 {
		t.Fatalf("expected score 77, got %d", entries[0].WokeScore)
	}
	if entries[0].Description != "mid woke" {
		t.Fatalf("expected 'mid woke', got %s", entries[0].Description)
	}
}

func TestAPIEntries_ContentTypeJSON(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/entries", nil)
	rr := httptest.NewRecorder()

	srv.handleAPIEntries(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestPhoto_PathTraversal(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/photos/../../../etc/passwd", nil)
	rr := httptest.NewRecorder()

	srv.handlePhoto(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for path traversal, got %d", rr.Code)
	}
}

func TestPhoto_ValidFile(t *testing.T) {
	srv := newTestServer(t)

	testFile := filepath.Join(srv.cfg.PhotoDir, "test123.jpg")
	os.WriteFile(testFile, []byte{0xFF, 0xD8, 0xFF, 0xD9}, 0644)

	req := httptest.NewRequest("GET", "/photos/test123.jpg", nil)
	rr := httptest.NewRecorder()

	srv.handlePhoto(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body, _ := io.ReadAll(rr.Body)
	if len(body) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(body))
	}
}

func TestPhoto_NotFound(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/photos/nonexistent.jpg", nil)
	rr := httptest.NewRecorder()

	srv.handlePhoto(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
