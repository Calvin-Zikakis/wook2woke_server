package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

const voterIDKey contextKey = "voterID"

func getVoterID(r *http.Request) string {
	if vid, ok := r.Context().Value(voterIDKey).(string); ok {
		return vid
	}
	return ""
}

func ensureVoterIDCookie(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("voter_id"); err == nil && c.Value != "" {
		return c.Value
	}
	vid := randomHex(16)
	http.SetCookie(w, &http.Cookie{
		Name:     "voter_id",
		Value:    vid,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
	})
	return vid
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	vid := getVoterID(r)
	if vid == "" {
		http.Error(w, "no voter id", http.StatusBadRequest)
		return
	}

	var body struct {
		Score int `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Score < 1 || body.Score > 7 {
		http.Error(w, "score must be 1-7", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(
		`INSERT INTO votes(entry_id, voter_id, score, created_at) VALUES(?,?,?,?)
		 ON CONFLICT(entry_id, voter_id) DO UPDATE SET score=excluded.score`,
		id, vid, body.Score, time.Now().UTC(),
	)
	if err != nil {
		log.Printf("vote: upsert failed for entry %d: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var avg float64
	var count int
	s.db.QueryRow("SELECT COALESCE(AVG(score),0), COUNT(*) FROM votes WHERE entry_id=?", id).Scan(&avg, &count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"avg":      avg,
		"count":    count,
		"userVote": body.Score,
	})
}
