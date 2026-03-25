package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockClaudeFn(score int, subject, description string) func(context.Context, []byte) (*claudeResult, error) {
	return func(_ context.Context, _ []byte) (*claudeResult, error) {
		return &claudeResult{Score: score, Subject: subject, Description: description}, nil
	}
}

func TestRescore_Success(t *testing.T) {
	srv := newTestServer(t)
	srv.claudeFn = mockClaudeFn(6, "human", "suspiciously well-hydrated festival goer")

	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 3, "original desc", "test.jpg")
	writeTestPhoto(t, srv.cfg.PhotoDir, "test.jpg")

	req := httptest.NewRequest("POST", "/api/entries/1/rescore", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	srv.handleRescore(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var rs Rescore
	json.NewDecoder(rr.Body).Decode(&rs)
	if rs.WokeScore != 6 {
		t.Fatalf("expected score 6, got %d", rs.WokeScore)
	}
	if rs.Subject != "human" {
		t.Fatalf("expected subject 'human', got %s", rs.Subject)
	}
	if rs.Description != "suspiciously well-hydrated festival goer" {
		t.Fatalf("unexpected description: %s", rs.Description)
	}
	if rs.EntryID != 1 {
		t.Fatalf("expected entry_id 1, got %d", rs.EntryID)
	}

	var count int
	srv.db.QueryRow("SELECT COUNT(*) FROM rescores").Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 rescore in db, got %d", count)
	}
}

func TestRescore_EntryNotFound(t *testing.T) {
	srv := newTestServer(t)
	srv.claudeFn = mockClaudeFn(5, "crystal", "pretty rock")

	req := httptest.NewRequest("POST", "/api/entries/99/rescore", nil)
	req.SetPathValue("id", "99")
	rr := httptest.NewRecorder()

	srv.handleRescore(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestRescore_InvalidID(t *testing.T) {
	srv := newTestServer(t)
	srv.claudeFn = mockClaudeFn(5, "crystal", "pretty rock")

	req := httptest.NewRequest("POST", "/api/entries/abc/rescore", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()

	srv.handleRescore(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRescore_MultipleRescores(t *testing.T) {
	srv := newTestServer(t)
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 4, "desc", "multi.jpg")
	writeTestPhoto(t, srv.cfg.PhotoDir, "multi.jpg")

	for i, score := range []int{2, 5, 7} {
		srv.claudeFn = mockClaudeFn(score, "human", "rescore "+string(rune('A'+i)))
		req := httptest.NewRequest("POST", "/api/entries/1/rescore", nil)
		req.SetPathValue("id", "1")
		rr := httptest.NewRecorder()
		srv.handleRescore(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("rescore %d failed: %d", i, rr.Code)
		}
	}

	var count int
	srv.db.QueryRow("SELECT COUNT(*) FROM rescores WHERE entry_id = 1").Scan(&count)
	if count != 3 {
		t.Fatalf("expected 3 rescores, got %d", count)
	}
}

func TestGetRescores_Empty(t *testing.T) {
	srv := newTestServer(t)
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 4, "desc", "photo.jpg")

	req := httptest.NewRequest("GET", "/api/entries/1/rescores", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	srv.handleGetRescores(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var rescores []Rescore
	json.NewDecoder(rr.Body).Decode(&rescores)
	if rescores != nil {
		t.Fatalf("expected nil, got %d rescores", len(rescores))
	}
}

func TestGetRescores_WithData(t *testing.T) {
	srv := newTestServer(t)
	srv.claudeFn = mockClaudeFn(7, "crystal", "very geometric indeed")
	srv.db.Exec("INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)", 3, "desc", "crystal.jpg")
	writeTestPhoto(t, srv.cfg.PhotoDir, "crystal.jpg")

	// Create a rescore via handleRescore
	req := httptest.NewRequest("POST", "/api/entries/1/rescore", nil)
	req.SetPathValue("id", "1")
	srv.handleRescore(httptest.NewRecorder(), req)

	// Fetch it back
	req = httptest.NewRequest("GET", "/api/entries/1/rescores", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	srv.handleGetRescores(rr, req)

	var rescores []Rescore
	json.NewDecoder(rr.Body).Decode(&rescores)
	if len(rescores) != 1 {
		t.Fatalf("expected 1 rescore, got %d", len(rescores))
	}
	if rescores[0].WokeScore != 7 {
		t.Fatalf("expected score 7, got %d", rescores[0].WokeScore)
	}
	if rescores[0].Subject != "crystal" {
		t.Fatalf("expected subject 'crystal', got %s", rescores[0].Subject)
	}
}

func TestGetRescores_InvalidID(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/entries/bad/rescores", nil)
	req.SetPathValue("id", "bad")
	rr := httptest.NewRecorder()

	srv.handleGetRescores(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
