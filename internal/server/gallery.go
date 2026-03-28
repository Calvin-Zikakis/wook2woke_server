package server

import (
	"crypto/subtle"
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

const entriesWithVotesSQL = `
SELECT e.id, e.woke_score, e.description, e.photo_path, e.created_at,
       COALESCE(AVG(v.score), 0.0), COALESCE(COUNT(v.id), 0),
       COALESCE(MAX(CASE WHEN v.voter_id = ? THEN v.score ELSE 0 END), 0)
FROM entries e
LEFT JOIN votes v ON v.entry_id = e.id
GROUP BY e.id
ORDER BY e.created_at DESC`

func (s *Server) handleGallery(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	vid := getVoterID(r)
	rows, err := s.db.Query(entriesWithVotesSQL, vid)
	if err != nil {
		log.Printf("gallery: db query failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.WokeScore, &e.Description, &e.PhotoPath, &e.CreatedAt, &e.VoteAvg, &e.VoteCount, &e.UserVote)
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
	vid := getVoterID(r)
	var e Entry
	err := s.db.QueryRow(`
		SELECT e.id, e.woke_score, e.description, e.photo_path, e.created_at,
		       COALESCE(AVG(v.score), 0.0), COALESCE(COUNT(v.id), 0),
		       COALESCE(MAX(CASE WHEN v.voter_id = ? THEN v.score ELSE 0 END), 0)
		FROM entries e
		LEFT JOIN votes v ON v.entry_id = e.id
		GROUP BY e.id
		ORDER BY e.created_at DESC LIMIT 1`, vid).
		Scan(&e.ID, &e.WokeScore, &e.Description, &e.PhotoPath, &e.CreatedAt, &e.VoteAvg, &e.VoteCount, &e.UserVote)
	if err != nil {
		liveTmpl.Execute(w, nil)
		return
	}
	liveTmpl.Execute(w, e)
}

func (s *Server) handleDeleteAll(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-API-Key")
	if subtle.ConstantTimeCompare([]byte(key), []byte(s.cfg.APIKey)) != 1 {
		log.Printf("delete-all: rejected request with bad API key from %s", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := s.db.Query("SELECT photo_path FROM entries")
	if err != nil {
		log.Printf("delete-all: failed to query entries: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var photoPaths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			photoPaths = append(photoPaths, p)
		}
	}

	if _, err := s.db.Exec("DELETE FROM rescores"); err != nil {
		log.Printf("delete-all: failed to delete rescores: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := s.db.Exec("DELETE FROM votes"); err != nil {
		log.Printf("delete-all: failed to delete votes: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := s.db.Exec("DELETE FROM entries"); err != nil {
		log.Printf("delete-all: failed to delete entries: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	for _, p := range photoPaths {
		if err := os.Remove(filepath.Join(s.cfg.PhotoDir, p)); err != nil {
			log.Printf("delete-all: failed to remove photo file %s: %v", p, err)
		}
	}

	log.Printf("delete-all: removed %d entries", len(photoPaths))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"deleted": len(photoPaths), "status": "ok"})
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
	vid := getVoterID(r)
	rows, err := s.db.Query(entriesWithVotesSQL, vid)
	if err != nil {
		log.Printf("api/entries: db query failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.WokeScore, &e.Description, &e.PhotoPath, &e.CreatedAt, &e.VoteAvg, &e.VoteCount, &e.UserVote)
		entries = append(entries, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
