package server

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"time"
)

type contextKey string

const roleKey contextKey = "role"

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	loginTmpl.Execute(w, nil)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	var role string
	if subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.AdminPassword)) == 1 {
		role = "admin"
	} else if subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.ViewPassword)) == 1 {
		role = "viewer"
	} else {
		log.Printf("auth: failed login attempt from %s", r.RemoteAddr)
		loginTmpl.Execute(w, map[string]string{"Error": "Wrong password"})
		return
	}

	token := randomHex(32)
	s.mu.Lock()
	s.sessions[token] = sessionData{expiry: time.Now().Add(s.cfg.SessionTTL), role: role}
	s.mu.Unlock()

	log.Printf("auth: successful login as %s from %s", role, r.RemoteAddr)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vid := ensureVoterIDCookie(w, r)

		// If no view password is set, allow public access but still honour admin sessions.
		if s.cfg.ViewPassword == "" {
			role := "viewer"
			if cookie, err := r.Cookie("session"); err == nil {
				s.mu.RLock()
				sd, ok := s.sessions[cookie.Value]
				s.mu.RUnlock()
				if ok && time.Now().Before(sd.expiry) {
					role = sd.role
				}
			}
			ctx := context.WithValue(r.Context(), roleKey, role)
			ctx = context.WithValue(ctx, voterIDKey, vid)
			next(w, r.WithContext(ctx))
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		s.mu.RLock()
		sd, ok := s.sessions[cookie.Value]
		s.mu.RUnlock()

		if !ok || time.Now().After(sd.expiry) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), roleKey, sd.role)
		ctx = context.WithValue(ctx, voterIDKey, vid)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(roleKey) != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}

func isAdmin(r *http.Request) bool {
	return r.Context().Value(roleKey) == "admin"
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		s.mu.Lock()
		delete(s.sessions, cookie.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	if s.cfg.ViewPassword == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
