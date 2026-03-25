package server

import (
	"crypto/subtle"
	"log"
	"net/http"
	"time"
)

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	loginTmpl.Execute(w, nil)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.ViewPassword)) != 1 {
		log.Printf("auth: failed login attempt from %s", r.RemoteAddr)
		loginTmpl.Execute(w, map[string]string{"Error": "Wrong password"})
		return
	}

	token := randomHex(32)
	s.mu.Lock()
	s.sessions[token] = time.Now().Add(s.cfg.SessionTTL)
	s.mu.Unlock()

	log.Printf("auth: successful login from %s", r.RemoteAddr)
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
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		s.mu.RLock()
		expiry, ok := s.sessions[cookie.Value]
		s.mu.RUnlock()

		if !ok || time.Now().After(expiry) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
