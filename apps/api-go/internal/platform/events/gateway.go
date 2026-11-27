package events

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/orbito/api-go/internal/middleware"
)

// Gateway implements an Engine.IO v4 / Socket.IO v4 real-time server for the /events namespace.
type Gateway struct {
	tokenValidator middleware.TokenValidator
	userChecker    middleware.UserChecker
	corsOrigins    []string
	log            *slog.Logger
	upgrader       websocket.Upgrader

	mu             sync.RWMutex
	sessions       map[string]*eioSession
	clients        map[string]*eioSession
	connectedUsers map[string]map[string]struct{}
	userLastSeen   map[string]string
	rooms          map[string]map[string]*eioSession
}

// NewGateway constructs a Socket.IO v4 gateway validating handshakes via TokenValidator and UserChecker.
func NewGateway(validator middleware.TokenValidator, checker middleware.UserChecker, corsOrigins []string, log *slog.Logger) *Gateway {
	gw := &Gateway{
		tokenValidator: validator,
		userChecker:    checker,
		corsOrigins:    corsOrigins,
		log:            log,
		sessions:       make(map[string]*eioSession),
		clients:        make(map[string]*eioSession),
		connectedUsers: make(map[string]map[string]struct{}),
		userLastSeen:   make(map[string]string),
		rooms:          make(map[string]map[string]*eioSession),
	}

	gw.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     gw.checkOrigin,
	}

	return gw
}

func (g *Gateway) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" || len(g.corsOrigins) == 0 {
		return true
	}
	for _, allowed := range g.corsOrigins {
		trimmed := strings.TrimSpace(allowed)
		if trimmed == "*" || strings.EqualFold(trimmed, origin) {
			return true
		}
	}
	return false
}

// ServeHTTP handles Engine.IO v4 polling and WebSocket requests under /socket.io/.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	transport := q.Get("transport")
	sid := q.Get("sid")

	switch transport {
	case "websocket":
		g.handleWebSocket(w, r, sid)
	case "polling", "":
		switch r.Method {
		case http.MethodGet:
			g.handlePollingGet(w, r, sid)
		case http.MethodPost:
			g.handlePollingPost(w, r, sid)
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":0,"message":"Transport unknown"}`))
	}
}

// Close terminates all active Socket.IO sessions during server shutdown.
func (g *Gateway) Close() {
	g.mu.RLock()
	active := make([]*eioSession, 0, len(g.sessions))
	for _, s := range g.sessions {
		active = append(active, s)
	}
	g.mu.RUnlock()

	for _, s := range active {
		s.close()
	}
}

func (g *Gateway) storeSession(sess *eioSession) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessions[sess.sid] = sess
}

func (g *Gateway) getSession(sid string) *eioSession {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.sessions[sid]
}

func (g *Gateway) removeSession(sid string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.sessions, sid)
}

