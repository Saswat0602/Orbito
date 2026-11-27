package events

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	namespaceEvents   = "/events"
	pingIntervalMS    = 25000
	pingTimeoutMS     = 20000
	maxPayloadBytes   = 1000000
	pollingWaitWindow = 10 * time.Second
	pollingSeparator  = "\x1e"
)

type eioHandshake struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
	MaxPayload   int      `json:"maxPayload"`
}

type eioSession struct {
	sid       string
	nsSID     string
	meta      handshakeMeta
	gateway   *Gateway
	outbox    chan string
	notifyCh  chan struct{}
	upgradeCh chan struct{}
	closeCh   chan struct{}
	closeOnce sync.Once

	wsMu sync.Mutex
	ws   *websocket.Conn

	userID         string
	organizationID string
	workspaceID    string
	projectID      string
	rooms          map[string]struct{}
}

func newSessionID() string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "orbito-" + time.Now().UTC().Format("150405.000")
	}
	return hex.EncodeToString(b)
}

func newEIOSession(g *Gateway, meta handshakeMeta) *eioSession {
	return &eioSession{
		sid:       newSessionID(),
		meta:      meta,
		gateway:   g,
		outbox:    make(chan string, 256),
		notifyCh:  make(chan struct{}, 1),
		upgradeCh: make(chan struct{}),
		closeCh:   make(chan struct{}),
		rooms:     make(map[string]struct{}),
	}
}

func (s *eioSession) enqueue(pkt string) {
	select {
	case <-s.closeCh:
		return
	default:
	}

	select {
	case s.outbox <- pkt:
		select {
		case s.notifyCh <- struct{}{}:
		default:
		}
	default:
		// Non-blocking drop if slow client overflows buffer
		if s.gateway.log != nil {
			gLog := s.gateway.log
			gLog.Warn("dropping socket packet due to full outbox", slog.String("sid", s.sid))
		}
	}
}

func (s *eioSession) drainOutbox() []string {
	var packets []string
	for {
		select {
		case pkt := <-s.outbox:
			packets = append(packets, pkt)
		default:
			return packets
		}
	}
}

func (s *eioSession) close() {
	s.closeOnce.Do(func() {
		close(s.closeCh)
		s.gateway.removeSession(s.sid)
		s.gateway.unregisterClient(s)
		s.wsMu.Lock()
		if s.ws != nil {
			_ = s.ws.Close()
		}
		s.wsMu.Unlock()
	})
}

func (g *Gateway) handlePollingGet(w http.ResponseWriter, r *http.Request, sid string) {
	if sid == "" {
		meta := extractHandshakeMeta(r)
		sess := newEIOSession(g, meta)
		g.storeSession(sess)

		hsBytes, err := json.Marshal(eioHandshake{
			SID:          sess.sid,
			Upgrades:     []string{"websocket"},
			PingInterval: pingIntervalMS,
			PingTimeout:  pingTimeoutMS,
			MaxPayload:   maxPayloadBytes,
		})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("0" + string(hsBytes)))
		return
	}

	sess := g.getSession(sid)
	if sess == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":1,"message":"Session ID unknown"}`))
		return
	}

	packets := sess.drainOutbox()
	if len(packets) == 0 {
		timer := time.NewTimer(pollingWaitWindow)
		defer timer.Stop()
		select {
		case <-sess.notifyCh:
			packets = sess.drainOutbox()
		case <-sess.upgradeCh:
			packets = []string{"6"}
		case <-sess.closeCh:
			packets = []string{"1"}
		case <-r.Context().Done():
			return
		case <-timer.C:
			packets = []string{"6"}
		}
	}

	if len(packets) == 0 {
		packets = []string{"6"}
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strings.Join(packets, pollingSeparator)))
}

func (g *Gateway) handlePollingPost(w http.ResponseWriter, r *http.Request, sid string) {
	sess := g.getSession(sid)
	if sess == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":1,"message":"Session ID unknown"}`))
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, maxPayloadBytes))
	if err != nil {
		http.Error(w, "failed to read payload", http.StatusBadRequest)
		return
	}

	parts := strings.Split(string(bodyBytes), pollingSeparator)
	for _, pkt := range parts {
		if pkt == "" {
			continue
		}
		g.processEIOPacket(r.Context(), sess, pkt, false)
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (g *Gateway) handleWebSocket(w http.ResponseWriter, r *http.Request, sid string) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	var sess *eioSession
	isUpgrade := false
	if sid != "" {
		sess = g.getSession(sid)
		if sess == nil {
			_ = conn.Close()
			return
		}
		isUpgrade = true
	} else {
		sess = newEIOSession(g, extractHandshakeMeta(r))
		g.storeSession(sess)
	}

	sess.wsMu.Lock()
	sess.ws = conn
	sess.wsMu.Unlock()

	if !isUpgrade {
		hsBytes, mErr := json.Marshal(eioHandshake{
			SID:          sess.sid,
			Upgrades:     []string{},
			PingInterval: pingIntervalMS,
			PingTimeout:  pingTimeoutMS,
			MaxPayload:   maxPayloadBytes,
		})
		if mErr == nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("0"+string(hsBytes)))
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.wsWriteLoop(sess, conn, isUpgrade)
	}()

	g.wsReadLoop(r.Context(), sess, conn)
	sess.close()
	wg.Wait()
}

func (g *Gateway) wsWriteLoop(sess *eioSession, conn *websocket.Conn, waitUpgrade bool) {
	pingTicker := time.NewTicker(time.Duration(pingIntervalMS) * time.Millisecond)
	defer pingTicker.Stop()

	upgraded := !waitUpgrade
	for {
		if !upgraded {
			select {
			case <-sess.upgradeCh:
				upgraded = true
			case <-sess.closeCh:
				return
			}
		}

		select {
		case pkt := <-sess.outbox:
			if err := writeWSMessage(conn, pkt); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := writeWSMessage(conn, "2"); err != nil {
				return
			}
		case <-sess.closeCh:
			return
		}
	}
}

func writeWSMessage(conn *websocket.Conn, msg string) error {
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (g *Gateway) wsReadLoop(ctx context.Context, sess *eioSession, conn *websocket.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(pingIntervalMS+pingTimeoutMS) * time.Millisecond))
	for {
		msgType, msgBytes, err := conn.ReadMessage()
		if err != nil {
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(pingIntervalMS+pingTimeoutMS) * time.Millisecond))
		if msgType != websocket.TextMessage {
			continue
		}
		pkt := string(msgBytes)
		if pkt == "" {
			continue
		}
		if pkt == "2probe" {
			_ = writeWSMessage(conn, "3probe")
			continue
		}
		if pkt == "5" {
			close(sess.upgradeCh)
			continue
		}
		g.processEIOPacket(ctx, sess, pkt, true)
	}
}

func (g *Gateway) processEIOPacket(ctx context.Context, sess *eioSession, pkt string, isWS bool) {
	if pkt == "" {
		return
	}
	switch pkt[0] {
	case '1':
		sess.close()
	case '2':
		if isWS {
			sess.enqueue("3")
		} else {
			sess.enqueue("3")
		}
	case '3':
		// Client pong received; read deadline already refreshed
	case '4':
		if len(pkt) > 1 {
			g.processSocketIOPacket(ctx, sess, pkt[1:])
		}
	}
}

func (g *Gateway) processSocketIOPacket(ctx context.Context, sess *eioSession, payload string) {
	if payload == "" {
		return
	}

	pktType := payload[0]
	rest := payload[1:]

	ns := "/"
	if strings.HasPrefix(rest, "/") {
		if idx := strings.IndexByte(rest, ','); idx >= 0 {
			ns = rest[:idx]
			rest = rest[idx+1:]
		} else {
			ns = rest
			rest = ""
		}
	} else if strings.HasPrefix(rest, ",") {
		rest = rest[1:]
	}

	if ns != namespaceEvents && ns != "/" {
		sess.enqueue("44" + ns + `,{"message":"Invalid namespace"}`)
		return
	}

	switch pktType {
	case '0':
		g.handleSocketIOConnect(ctx, sess, ns, rest)
	case '1':
		if ns == namespaceEvents {
			g.unregisterClient(sess)
		}
	case '2':
		if ns == namespaceEvents {
			g.handleSocketIOEvent(sess, rest)
		}
	}
}

func (g *Gateway) handleSocketIOConnect(ctx context.Context, sess *eioSession, ns, authBody string) {
	if ns != namespaceEvents {
		// Acknowledge root namespace connection quietly so multiplexed clients stay alive
		sess.enqueue(`40{"sid":"` + newSessionID() + `"}`)
		return
	}

	var authObj struct {
		Token string `json:"token"`
	}
	if authBody != "" {
		_ = json.Unmarshal([]byte(authBody), &authObj)
	}

	tokenStr := resolveToken(authObj.Token, sess.meta)
	userID, err := g.authenticateToken(ctx, tokenStr)

	nsSID := newSessionID()
	sess.nsSID = nsSID
	sess.enqueue("40" + namespaceEvents + `,{"sid":"` + nsSID + `"}`)

	if err != nil {
		msg := "Authentication failed"
		details := err.Error()
		if details == "Authentication token required" {
			g.emitToSession(sess, "error", map[string]any{"message": details})
		} else {
			g.emitToSession(sess, "error", map[string]any{"message": msg, "details": details})
		}
		sess.enqueue("41" + namespaceEvents)
		return
	}

	wasOffline := g.registerAuthenticatedClient(sess, userID)
	g.emitToSession(sess, "connected", map[string]any{
		"userId":    userID,
		"socketId":  nsSID,
		"timestamp": nowISO(),
	})

	if wasOffline {
		g.broadcastAll("user:online", map[string]any{
			"userId":    userID,
			"isOnline":  true,
			"timestamp": nowISO(),
		})
	}
}

func (g *Gateway) handleSocketIOEvent(sess *eioSession, rest string) {
	idx := strings.IndexByte(rest, '[')
	if idx < 0 {
		return
	}
	ackID := strings.TrimSpace(rest[:idx])
	jsonArr := rest[idx:]

	var elems []json.RawMessage
	if err := json.Unmarshal([]byte(jsonArr), &elems); err != nil || len(elems) == 0 {
		return
	}

	var eventName string
	if err := json.Unmarshal(elems[0], &eventName); err != nil {
		return
	}

	var data json.RawMessage
	if len(elems) > 1 {
		data = elems[1]
	}

	g.handleClientEvent(sess, eventName, data)

	if ackID != "" {
		sess.enqueue("43" + namespaceEvents + "," + ackID + "[]")
	}
}
