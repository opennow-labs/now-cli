package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPushStatus(t *testing.T) {
	var received StatusRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/status" {
			t.Errorf("expected /api/status, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer now_test" {
			t.Errorf("missing or wrong Authorization header: %s", r.Header.Get("Authorization"))
		}

		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(200)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "now_test")
	err := client.PushStatus("coding in Go", "\U0001F4BB")
	if err != nil {
		t.Fatalf("PushStatus: %v", err)
	}
	if received.Content != "coding in Go" {
		t.Errorf("content = %q, want %q", received.Content, "coding in Go")
	}
	if received.Emoji != "\U0001F4BB" {
		t.Errorf("emoji = %q, want %q", received.Emoji, "\U0001F4BB")
	}
}

func TestVerifyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/me" {
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`{"user": {"id": 1, "name": "testuser"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "now_test")
	me, err := client.VerifyToken()
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if me.User.Name != "testuser" {
		t.Errorf("name = %q, want %q", me.User.Name, "testuser")
	}
	if me.User.ID != 1 {
		t.Errorf("id = %d, want 1", me.User.ID)
	}
}

func TestGetBoard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"board": [{"id": 1, "name": "alice", "type": "human", "status": "coding", "emoji": "💻"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "now_test")
	board, err := client.GetBoard()
	if err != nil {
		t.Fatalf("GetBoard: %v", err)
	}
	if len(board.Board) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(board.Board))
	}
	if board.Board[0].Name != "alice" {
		t.Errorf("name = %q, want %q", board.Board[0].Name, "alice")
	}
}

func TestRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "10")
		w.WriteHeader(429)
		w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "now_test")
	err := client.PushStatus("test", "")
	if err == nil {
		t.Fatal("expected error on 429")
	}

	rle, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if rle.RetryAfter != 10*time.Second {
		t.Errorf("RetryAfter = %v, want 10s", rle.RetryAfter)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad_token")
	err := client.PushStatus("test", "")
	if err == nil {
		t.Fatal("expected error on 401")
	}
	if err.Error() != "API error (401): invalid token" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("expected no auth header for empty token")
		}
		w.Write([]byte(`{"board": []}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.GetBoard()
	if err != nil {
		t.Fatalf("GetBoard: %v", err)
	}
}
