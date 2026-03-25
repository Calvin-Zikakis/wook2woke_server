package server

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-API-Key")
	if subtle.ConstantTimeCompare([]byte(key), []byte(s.cfg.APIKey)) != 1 {
		log.Printf("upload: rejected request with bad API key from %s", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		log.Printf("upload: failed to parse multipart form: %v", err)
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	wokeScoreStr := r.FormValue("wokeScore")
	wokeScore, err := strconv.Atoi(wokeScoreStr)
	if err != nil {
		log.Printf("upload: invalid wokeScore %q: %v", wokeScoreStr, err)
		http.Error(w, "wokeScore must be an integer", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		log.Printf("upload: missing description")
		http.Error(w, "description is required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		log.Printf("upload: missing photo field: %v", err)
		http.Error(w, "photo is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(header.Filename), ".jpeg") {
		ct := header.Header.Get("Content-Type")
		if ct != "image/jpeg" {
			log.Printf("upload: rejected non-JPEG file %q (content-type: %s)", header.Filename, ct)
			http.Error(w, "photo must be a JPEG", http.StatusBadRequest)
			return
		}
	}

	filename := fmt.Sprintf("%d_%s.jpg", time.Now().UnixNano(), randomHex(8))
	destPath := filepath.Join(s.cfg.PhotoDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		log.Printf("upload: failed to create file %s: %v", destPath, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	n, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("upload: failed to write file %s: %v", destPath, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("upload: saved photo %s (%d bytes)", filename, n)

	result, err := s.db.Exec(
		"INSERT INTO entries (woke_score, description, photo_path) VALUES (?, ?, ?)",
		wokeScore, description, filename,
	)
	if err != nil {
		log.Printf("upload: db insert failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	log.Printf("upload: created entry id=%d score=%d desc=%q", id, wokeScore, description)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"id": id, "status": "ok"})
}
