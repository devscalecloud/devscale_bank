package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{name: "positive", input: 25430.5, want: "$25,430.50"},
		{name: "negative", input: -1200.25, want: "-$1,200.25"},
		{name: "zero", input: 0, want: "$0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCurrency(tt.input)
			if got != tt.want {
				t.Fatalf("formatCurrency(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLookupHandlerUnauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/lookup?account=0987654321", nil)
	rr := httptest.NewRecorder()

	lookupHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestLookupHandlerValidAccount(t *testing.T) {
	token := "test-token"
	mu.Lock()
	sessions[token] = Session{Username: "dapo", ExpiresAt: time.Now().Add(time.Hour)}
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(sessions, token)
		mu.Unlock()
	}()

	req := httptest.NewRequest(http.MethodGet, "/lookup?account=0987654321", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()

	lookupHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); body != "Jane Smith" {
		t.Fatalf("body = %q, want %q", body, "Jane Smith")
	}
}
