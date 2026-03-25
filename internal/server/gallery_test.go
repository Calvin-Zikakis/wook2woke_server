package server

import (
	"bytes"
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

func TestDelete_Success(t *testing.T) {
	srv := newTestServer(t)
	writeTestPhoto(t, srv.cfg.PhotoDir, "del.jpg")
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 3, "to delete", "del.jpg")

	req := httptest.NewRequest("DELETE", "/api/entries/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	srv.handleDelete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	var count int
	srv.db.QueryRow("SELECT COUNT(*) FROM entries").Scan(&count)
	if count != 0 {
		t.Fatalf("expected 0 entries after delete, got %d", count)
	}

	if _, err := os.Stat(filepath.Join(srv.cfg.PhotoDir, "del.jpg")); !os.IsNotExist(err) {
		t.Fatal("expected photo file to be deleted")
	}
}

func TestDelete_NotFound(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("DELETE", "/api/entries/99", nil)
	req.SetPathValue("id", "99")
	rr := httptest.NewRecorder()
	srv.handleDelete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDelete_InvalidID(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("DELETE", "/api/entries/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	srv.handleDelete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDelete_AlsoRemovesRescores(t *testing.T) {
	srv := newTestServer(t)
	writeTestPhoto(t, srv.cfg.PhotoDir, "with-rescores.jpg")
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 4, "desc", "with-rescores.jpg")
	srv.db.Exec("INSERT INTO rescores (entry_id, woke_score, subject, description) VALUES (1, 5, 'human', 'rescored')")
	srv.db.Exec("INSERT INTO rescores (entry_id, woke_score, subject, description) VALUES (1, 6, 'human', 'rescored again')")

	req := httptest.NewRequest("DELETE", "/api/entries/1", nil)
	req.SetPathValue("id", "1")
	srv.handleDelete(httptest.NewRecorder(), req)

	var count int
	srv.db.QueryRow("SELECT COUNT(*) FROM rescores WHERE entry_id = 1").Scan(&count)
	if count != 0 {
		t.Fatalf("expected rescores to be deleted, got %d", count)
	}
}

func TestPromoteRescore_Success(t *testing.T) {
	srv := newTestServer(t)
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 3, "original", "p.jpg")

	body := bytes.NewBufferString(`{"wokeScore":6,"description":"promoted desc"}`)
	req := httptest.NewRequest("PUT", "/api/entries/1", body)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	srv.handlePromoteRescore(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	var score int
	var desc string
	srv.db.QueryRow("SELECT woke_score, description FROM entries WHERE id = 1").Scan(&score, &desc)
	if score != 6 {
		t.Fatalf("expected score 6 after promote, got %d", score)
	}
	if desc != "promoted desc" {
		t.Fatalf("expected 'promoted desc', got %s", desc)
	}
}

func TestPromoteRescore_InvalidID(t *testing.T) {
	srv := newTestServer(t)

	body := bytes.NewBufferString(`{"wokeScore":5,"description":"test"}`)
	req := httptest.NewRequest("PUT", "/api/entries/abc", body)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	srv.handlePromoteRescore(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPromoteRescore_InvalidBody(t *testing.T) {
	srv := newTestServer(t)
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 3, "original", "p.jpg")

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest("PUT", "/api/entries/1", body)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	srv.handlePromoteRescore(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
