package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginPage(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/login", nil)
	rr := httptest.NewRecorder()

	srv.handleLoginPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "wook2woke") {
		t.Fatal("login page should contain title")
	}
}

func TestLogin_Success(t *testing.T) {
	srv := newTestServer(t)

	form := url.Values{"password": {"test-password"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	srv.handleLogin(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to /, got %s", rr.Header().Get("Location"))
	}

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}

	srv.mu.RLock()
	if len(srv.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(srv.sessions))
	}
	srv.mu.RUnlock()
}

func TestLogin_WrongPassword(t *testing.T) {
	srv := newTestServer(t)

	form := url.Values{"password": {"wrong"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	srv.handleLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 (re-render login), got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Wrong password") {
		t.Fatal("expected error message in response")
	}
}

func TestLogin_EmptyPassword(t *testing.T) {
	srv := newTestServer(t)

	form := url.Values{"password": {""}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	srv.handleLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 (re-render login), got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Wrong password") {
		t.Fatal("expected error message for empty password")
	}
}

func TestRequireAuth_NoSession(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Fatalf("expected redirect to /login, got %s", rr.Header().Get("Location"))
	}
}

func TestRequireAuth_InvalidSession(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bogus-token"})
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rr.Code)
	}
}

func TestRequireAuth_ExpiredSession(t *testing.T) {
	srv := newTestServer(t)

	srv.mu.Lock()
	srv.sessions["expired-token"] = time.Now().Add(-time.Hour)
	srv.mu.Unlock()

	handler := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "expired-token"})
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect for expired session, got %d", rr.Code)
	}
}

func TestRequireAuth_ValidSession(t *testing.T) {
	srv := newTestServer(t)

	srv.mu.Lock()
	srv.sessions["valid-token"] = time.Now().Add(time.Hour)
	srv.mu.Unlock()

	called := false
	handler := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-token"})
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !called {
		t.Fatal("expected handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
