package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIntegration_UploadThenView(t *testing.T) {
	srv := newTestServer(t)

	// 1. Upload via ESP
	req := newUploadRequest(t, "test-api-key", 85, "pretty woke individual", "capture.jpg")
	rr := httptest.NewRecorder()
	srv.handleUpload(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("upload failed: %d %s", rr.Code, rr.Body.String())
	}

	// 2. Login
	form := url.Values{"password": {"test-password"}}
	req = httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	srv.handleLogin(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("login failed: %d", rr.Code)
	}

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie after login")
	}

	// 3. View gallery
	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	srv.requireAuth(srv.handleGallery)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("gallery failed: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "pretty woke individual") {
		t.Fatal("gallery should contain uploaded entry")
	}
	if !strings.Contains(rr.Body.String(), "85") {
		t.Fatal("gallery should contain woke score")
	}

	// 4. View JSON API
	req = httptest.NewRequest("GET", "/api/entries", nil)
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	srv.requireAuth(srv.handleAPIEntries)(rr, req)

	var entries []Entry
	json.NewDecoder(rr.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].WokeScore != 85 {
		t.Fatalf("expected score 85, got %d", entries[0].WokeScore)
	}
}
