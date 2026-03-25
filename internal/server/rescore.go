package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Rescore struct {
	ID          int    `json:"ID"`
	EntryID     int    `json:"EntryID"`
	WokeScore   int    `json:"WokeScore"`
	Subject     string `json:"Subject"`
	Description string `json:"Description"`
	CreatedAt   string `json:"CreatedAt"`
}

type claudeResult struct {
	Score       int    `json:"score"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

const rescorePrompt = `You are the analyst for "Wook or Woke," an art installation. ` +
	`You will see either a crystal or a human. Rate them 1-7 on the wook-to-woke spectrum: ` +
	`1=max wook (raw, earthy, chaotic, festival energy), 7=max woke (polished, geometric, precise, museum-ready). ` +
	`For crystals: color (warm=wook, cool=woke), clarity (cloudy=wook, clear=woke), ` +
	`shape (irregular=wook, geometric=woke), surface (rough=wook, polished=woke). ` +
	`For humans: vibe, outfit, hair, jewelry, accessories, energy (tie-dye/dreads/crystals=wook, ` +
	`minimalist/clean-cut/techwear=woke). ` +
	`If the image is neither a crystal nor a human, score 0. ` +
	`Respond ONLY with raw JSON, no markdown, no code blocks: ` +
	`{"score":<0-7>,"subject":"crystal" or "human" or "unknown","description":"<playful, max 80 chars>"}`

func (s *Server) handleRescore(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("rescore: invalid id %q", idStr)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var photoPath string
	if err := s.db.QueryRow("SELECT photo_path FROM entries WHERE id = ?", id).Scan(&photoPath); err != nil {
		log.Printf("rescore: entry %d not found: %v", id, err)
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	log.Printf("rescore: entry %d photo=%s", id, photoPath)

	imgData, err := os.ReadFile(filepath.Join(s.cfg.PhotoDir, photoPath))
	if err != nil {
		log.Printf("rescore: failed to read photo %s: %v", photoPath, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("rescore: sending %d bytes to Claude for entry %d", len(imgData), id)

	result, err := s.callClaude(r.Context(), imgData)
	if err != nil {
		log.Printf("rescore: claude error for entry %d: %v", id, err)
		http.Error(w, "analysis failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("rescore: entry %d result score=%d subject=%s desc=%q", id, result.Score, result.Subject, result.Description)

	res, err := s.db.Exec(
		"INSERT INTO rescores (entry_id, woke_score, subject, description) VALUES (?, ?, ?, ?)",
		id, result.Score, result.Subject, result.Description,
	)
	if err != nil {
		log.Printf("rescore: db insert failed for entry %d: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rescoreID, _ := res.LastInsertId()
	log.Printf("rescore: saved rescore %d for entry %d", rescoreID, id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Rescore{
		ID:          int(rescoreID),
		EntryID:     id,
		WokeScore:   result.Score,
		Subject:     result.Subject,
		Description: result.Description,
	})
}

func (s *Server) handleGetRescores(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	rows, err := s.db.Query(
		"SELECT id, entry_id, woke_score, subject, description, created_at FROM rescores WHERE entry_id = ? ORDER BY created_at DESC",
		id,
	)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rescores []Rescore
	for rows.Next() {
		var rs Rescore
		rows.Scan(&rs.ID, &rs.EntryID, &rs.WokeScore, &rs.Subject, &rs.Description, &rs.CreatedAt)
		rescores = append(rescores, rs)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rescores)
}

func (s *Server) callClaude(ctx context.Context, imgData []byte) (*claudeResult, error) {
	if s.claudeFn != nil {
		return s.claudeFn(ctx, imgData)
	}

	log.Printf("claude: creating client, sending %d byte image", len(imgData))
	client := anthropic.NewClient(option.WithAPIKey(s.cfg.AnthropicAPIKey))
	imgB64 := base64.StdEncoding.EncodeToString(imgData)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64("image/jpeg", imgB64),
				anthropic.NewTextBlock(rescorePrompt),
			),
		},
	})
	if err != nil {
		log.Printf("claude: API error: %v", err)
		return nil, err
	}

	text := msg.Content[0].AsText().Text
	log.Printf("claude: raw response: %s", text)

	// Strip markdown code fences if Claude ignores the prompt instructions
	clean := strings.TrimSpace(text)
	if strings.HasPrefix(clean, "```") {
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		clean = strings.TrimSpace(clean)
	}

	var result claudeResult
	if err := json.Unmarshal([]byte(clean), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (raw: %s)", err, text)
	}

	return &result, nil
}
