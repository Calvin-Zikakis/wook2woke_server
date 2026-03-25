package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUpload_Success(t *testing.T) {
	srv := newTestServer(t)
	req := newUploadRequest(t, "test-api-key", 42, "very woke person", "photo.jpg")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", resp["status"])
	}
	if resp["id"].(float64) != 1 {
		t.Fatalf("expected id 1, got %v", resp["id"])
	}

	var count int
	srv.db.QueryRow("SELECT COUNT(*) FROM entries").Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 entry in db, got %d", count)
	}

	files, _ := os.ReadDir(srv.cfg.PhotoDir)
	if len(files) != 1 {
		t.Fatalf("expected 1 photo file, got %d", len(files))
	}
}

func TestUpload_BadAPIKey(t *testing.T) {
	srv := newTestServer(t)
	req := newUploadRequest(t, "wrong-key", 42, "desc", "photo.jpg")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpload_NoAPIKey(t *testing.T) {
	srv := newTestServer(t)
	req := newUploadRequest(t, "", 42, "desc", "photo.jpg")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpload_MissingDescription(t *testing.T) {
	srv := newTestServer(t)
	req := newUploadRequest(t, "test-api-key", 42, "", "photo.jpg")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_NonJPEG(t *testing.T) {
	srv := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("wokeScore", "42")
	w.WriteField("description", "test")
	part, _ := w.CreateFormFile("photo", "photo.png")
	part.Write([]byte{0x89, 0x50, 0x4E, 0x47}) // PNG magic bytes
	w.Close()

	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-API-Key", "test-api-key")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "JPEG") {
		t.Fatalf("expected JPEG error message, got: %s", rr.Body.String())
	}
}

func TestUpload_InvalidWokeScore(t *testing.T) {
	srv := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("wokeScore", "not-a-number")
	w.WriteField("description", "test")
	part, _ := w.CreateFormFile("photo", "photo.jpg")
	part.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
	w.Close()

	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-API-Key", "test-api-key")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpload_MissingPhoto(t *testing.T) {
	srv := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("wokeScore", "42")
	w.WriteField("description", "test")
	w.Close()

	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-API-Key", "test-api-key")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpload_ContentTypeJPEG(t *testing.T) {
	srv := newTestServer(t)

	// File has no .jpg extension but declares image/jpeg content-type
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("wokeScore", "5")
	w.WriteField("description", "content-type accepted")
	fw, _ := w.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="photo"; filename="capture"`},
		"Content-Type":        {"image/jpeg"},
	})
	fw.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
	w.Close()

	req := httptest.NewRequest("POST", "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-API-Key", "test-api-key")
	rr := httptest.NewRecorder()

	srv.handleUpload(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for image/jpeg content-type, got %d: %s", rr.Code, rr.Body.String())
	}
}
