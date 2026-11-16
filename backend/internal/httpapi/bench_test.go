package httpapi_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/httpapi"
)

// The benchmarks measure the routes that read a whole progression at the most
// slots the API takes, through the handler. They run only with -bench; make
// check does not.
func benchmarkRoute(b *testing.B, path, body string) {
	b.Helper()
	transitions, err := catalog.LoadFunctionTransitions()
	if err != nil {
		b.Fatalf("loading the catalogue: %v", err)
	}
	handler := httpapi.NewHandler(httpapi.Deps{Logger: slog.New(slog.DiscardHandler), Transitions: transitions})
	b.ReportAllocs()
	for b.Loop() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(b.Context(), http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("status = %d\n%s", rec.Code, rec.Body)
		}
	}
}

func BenchmarkReadings256Slots(b *testing.B) {
	benchmarkRoute(b, readingsPath, slotsBody(256))
}

func BenchmarkNextCandidates256Slots(b *testing.B) {
	benchmarkRoute(b, nextCandidatesPath, `{"progression": `+slotsBody(256)+`, "beats": 1}`)
}

func BenchmarkSlotCandidates256Slots(b *testing.B) {
	benchmarkRoute(b, slotCandidatesPath, `{"progression": `+slotsBody(256)+`, "slot": 128}`)
}
