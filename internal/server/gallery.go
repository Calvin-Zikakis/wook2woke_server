package server

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func (s *Server) handleGallery(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	rows, err := s.db.Query("SELECT id, woke_score, description, photo_path, created_at FROM entries ORDER BY created_at DESC")
	if err != nil {
		log.Printf("gallery: db query failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.WokeScore, &e.Description, &e.PhotoPath, &e.CreatedAt)
		entries = append(entries, e)
	}

	galleryTmpl.Execute(w, entries)
}

func (s *Server) handlePhoto(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/photos/")
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' || c == '-') {
			log.Printf("gallery: rejected photo path traversal attempt: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
	}
	http.ServeFile(w, r, filepath.Join(s.cfg.PhotoDir, name))
}

func (s *Server) handleAPIEntries(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, woke_score, description, photo_path, created_at FROM entries ORDER BY created_at DESC")
	if err != nil {
		log.Printf("api/entries: db query failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.WokeScore, &e.Description, &e.PhotoPath, &e.CreatedAt)
		entries = append(entries, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
