// Package httpapi holds the HTTP handlers. A handler decodes the request,
// calls the engine and encodes the answer; it holds no domain logic.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// APIPrefix is the path every operation of the contract lives under; it is
// the server URL of docs/api/openapi.yaml.
const APIPrefix = "/api/v1"

// Version is the build's version, set at link time (see the Makefile).
var Version = "dev"

// Deps are the things the handlers need from outside.
type Deps struct {
	Logger *slog.Logger
}

// route is one operation of the API: a method, a path below APIPrefix and
// its handler.
type route struct {
	method  string
	path    string
	handler func(*api) http.HandlerFunc
}

// routes lists every operation the server offers. The contract lists the
// same operations; a test keeps the two in step.
var routes = []route{
	{method: http.MethodGet, path: "/health", handler: (*api).health},
}

type api struct {
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewHandler returns the API's handler with every route registered.
func NewHandler(deps Deps) http.Handler {
	a := &api{logger: deps.Logger, mux: http.NewServeMux()}
	for _, r := range routes {
		a.mux.HandleFunc(r.method+" "+APIPrefix+r.path, r.handler(a))
	}
	return a
}

// ServeHTTP answers a request no route matches with a problem document
// instead of the mux's plain text, keeping its status and Allow header.
func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, pattern := a.mux.Handler(r)
	if pattern != "" {
		handler.ServeHTTP(w, r)
		return
	}

	probe := &statusProbe{header: http.Header{}, status: http.StatusOK}
	handler.ServeHTTP(probe, r)
	if allow := probe.header.Get("Allow"); allow != "" {
		w.Header().Set("Allow", allow)
	}
	a.writeProblem(w, probe.status, "")
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (a *api) health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		a.writeJSON(w, http.StatusOK, "application/json", healthResponse{Status: "ok", Version: Version})
	}
}

// problem is an RFC 9457 problem document.
type problem struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func (a *api) writeProblem(w http.ResponseWriter, status int, detail string) {
	a.writeJSON(w, status, "application/problem+json",
		problem{Title: http.StatusText(status), Status: status, Detail: detail})
}

func (a *api) writeJSON(w http.ResponseWriter, status int, contentType string, body any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		a.logger.Error("encoding the response failed", "error", err)
	}
}

// statusProbe records the status and headers a handler writes and drops the body.
type statusProbe struct {
	header http.Header
	status int
}

func (p *statusProbe) Header() http.Header         { return p.header }
func (p *statusProbe) Write(b []byte) (int, error) { return len(b), nil }
func (p *statusProbe) WriteHeader(status int)      { p.status = status }
