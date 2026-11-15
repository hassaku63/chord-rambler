package httpapi_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/httpapi"
)

// serve sends one request to the API's handler and returns what it wrote.
func serve(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	handler := httpapi.NewHandler(httpapi.Deps{Logger: slog.New(slog.DiscardHandler)})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, http.NoBody))
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var body T
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding the body: %v", err)
	}
	return body
}

func TestHealthReportsStatusAndVersion(t *testing.T) {
	rec := serve(t, http.MethodGet, "/api/v1/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	body := decode[struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}](t, rec)
	if body.Status != "ok" {
		t.Errorf("status field = %q, want %q", body.Status, "ok")
	}
	if body.Version != httpapi.Version {
		t.Errorf("version field = %q, want %q", body.Version, httpapi.Version)
	}
}

func TestUnmatchedRequestsGetAProblemDocument(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantAllow  string
	}{
		{name: "unknown path", method: http.MethodGet, path: "/api/v1/nothing", wantStatus: http.StatusNotFound},
		{name: "outside the API", method: http.MethodGet, path: "/", wantStatus: http.StatusNotFound},
		{name: "known path, other method", method: http.MethodPost, path: "/api/v1/health",
			wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, tt.method, tt.path)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/problem+json")
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
			body := decode[struct {
				Title  string `json:"title"`
				Status int    `json:"status"`
			}](t, rec)
			if body.Status != tt.wantStatus {
				t.Errorf("status field = %d, want %d", body.Status, tt.wantStatus)
			}
			if body.Title != http.StatusText(tt.wantStatus) {
				t.Errorf("title field = %q, want %q", body.Title, http.StatusText(tt.wantStatus))
			}
		})
	}
}
