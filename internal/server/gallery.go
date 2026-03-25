package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GalleryData struct {
	Entries []Entry
	IsAdmin bool
}

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

	galleryTmpl.Execute(w, GalleryData{Entries: entries, IsAdmin: isAdmin(r)})
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

func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	var e Entry
	err := s.db.QueryRow("SELECT id, woke_score, description, photo_path, created_at FROM entries ORDER BY created_at DESC LIMIT 1").
		Scan(&e.ID, &e.WokeScore, &e.Description, &e.PhotoPath, &e.CreatedAt)
	if err != nil {
		liveTmpl.Execute(w, nil)
		return
	}
	liveTmpl.Execute(w, e)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var photoPath string
	if err := s.db.QueryRow("SELECT photo_path FROM entries WHERE id = ?", id).Scan(&photoPath); err != nil {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	if _, err := s.db.Exec("DELETE FROM rescores WHERE entry_id = ?", id); err != nil {
		log.Printf("delete: failed to remove rescores for entry %d: %v", id, err)
	}
	if _, err := s.db.Exec("DELETE FROM entries WHERE id = ?", id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := os.Remove(filepath.Join(s.cfg.PhotoDir, photoPath)); err != nil {
		log.Printf("delete: failed to remove photo file %s: %v", photoPath, err)
	}

	log.Printf("delete: removed entry %d", id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePromoteRescore(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		WokeScore   int    `json:"wokeScore"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if _, err := s.db.Exec("UPDATE entries SET woke_score = ?, description = ? WHERE id = ?", body.WokeScore, body.Description, id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("promote: entry %d updated to score=%d", id, body.WokeScore)
	w.WriteHeader(http.StatusNoContent)
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
