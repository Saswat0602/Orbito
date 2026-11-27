package queue

import (
	"crypto/subtle"
	"encoding/base64"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/shared/httpx"
)

var boardPageTpl = template.Must(template.New("queue-board").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Orbito Queue Dashboard</title></head>
<body style="font-family:sans-serif;background:#09090b;color:#f4f4f5;padding:24px;">
<h1>Orbito Queue Monitor</h1>
<p>Backend: <strong>{{.Backend}}</strong> (Requested: {{.RequestedBackend}}, Redis Available: {{.RedisAvailable}})</p>
<p>Stats: Total={{.Stats.TotalJobs}}, Completed={{.Stats.CompletedJobs}}, Failed={{.Stats.FailedJobs}}, Active={{.Stats.ActiveJobs}}</p>
<ul>{{range .Queues}}<li>{{.}}</li>{{end}}</ul>
</body>
</html>`))

// HealthHandler returns an HTTP handler for GET /api/health/queue.
func HealthHandler(rt *Runtime) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.RespondJSON(w, http.StatusOK, rt.Health())
	}
}

// DashboardRoutes builds the /api/queues router.
func DashboardRoutes(rt *Runtime, bullBoardEmail, bullBoardPassword string) chi.Router {
	qr := chi.NewRouter()
	qr.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		httpx.RespondJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"message": "Queue UI Controller is reachable",
		})
	})
	qr.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/api/health/queue", http.StatusTemporaryRedirect)
	})

	dashboard := newBoardHandler(rt, bullBoardEmail, bullBoardPassword)
	qr.HandleFunc("/", dashboard)
	qr.HandleFunc("/*", dashboard)
	return qr
}

// MountRoutes mounts public queue health and Basic-Auth-protected queue UI endpoints.
func MountRoutes(r chi.Router, rt *Runtime, bullBoardEmail, bullBoardPassword string) {
	r.Get("/health/queue", HealthHandler(rt))
	r.Mount("/queues", DashboardRoutes(rt, bullBoardEmail, bullBoardPassword))
}

func newBoardHandler(rt *Runtime, expectedEmail, expectedPassword string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if expectedEmail != "" && expectedPassword != "" {
			if !authorizeBasicAuth(w, req, expectedEmail, expectedPassword) {
				return
			}
		}

		health := rt.Health()
		if strings.Contains(req.Header.Get("Accept"), "text/html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_ = boardPageTpl.Execute(w, health)
			return
		}
		httpx.RespondJSON(w, http.StatusOK, health)
	}
}

func authorizeBasicAuth(w http.ResponseWriter, req *http.Request, expectedEmail, expectedPassword string) bool {
	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Basic ") {
		writeUnauthorized(w, "Authentication required")
		return false
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
	if err != nil {
		writeUnauthorized(w, "Invalid credentials")
		return false
	}
	user, pass, ok := strings.Cut(string(raw), ":")
	if !ok {
		writeUnauthorized(w, "Invalid credentials")
		return false
	}
	userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(expectedEmail)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPassword)) == 1
	if !userMatch || !passMatch {
		writeUnauthorized(w, "Invalid credentials")
		return false
	}
	return true
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Bull Board"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(msg))
}
