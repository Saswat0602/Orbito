package events

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTokenValidator struct {
	tokens map[string]struct {
		sub   string
		email string
		role  string
	}
}

func (s *stubTokenValidator) ValidateToken(tokenStr string) (string, string, string, error) {
	rec, ok := s.tokens[tokenStr]
	if !ok {
		return "", "", "", assert.AnError
	}
	return rec.sub, rec.email, rec.role, nil
}

func dialWS(t *testing.T, srvURL string, headers http.Header, query string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(srvURL, "http") + "/socket.io/?EIO=4&transport=websocket" + query
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func readWSMsg(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	return string(msg)
}

func connectNamespace(t *testing.T, conn *websocket.Conn, connectBody string) {
	t.Helper()
	openPkt := readWSMsg(t, conn)
	require.True(t, strings.HasPrefix(openPkt, "0{"), "expected EIO open packet, got %s", openPkt)

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("40/events,"+connectBody)))
	ackPkt := readWSMsg(t, conn)
	require.True(t, strings.HasPrefix(ackPkt, "40/events,"), "expected SIO namespace connect ack, got %s", ackPkt)
}

func TestGateway_AuthAndPresence(t *testing.T) {
	user1 := uuid.New()
	user2 := uuid.New()
	suspendedUser := uuid.New()

	val := &stubTokenValidator{
		tokens: map[string]struct {
			sub   string
			email string
			role  string
		}{
			"tok-u1":   {sub: user1.String(), email: "u1@orbito.io", role: "MEMBER"},
			"tok-u2":   {sub: user2.String(), email: "u2@orbito.io", role: "MEMBER"},
			"tok-susp": {sub: suspendedUser.String(), email: "s@orbito.io", role: "MEMBER"},
		},
	}

	checker := func(_ context.Context, id uuid.UUID) (*sqlc.User, error) {
		if id == suspendedUser {
			return &sqlc.User{ID: id, Status: sqlc.UserStatusSUSPENDED}, nil
		}
		return &sqlc.User{ID: id, Status: sqlc.UserStatusACTIVE, DeletedAt: pgtype.Timestamp{Valid: false}}, nil
	}

	gw := NewGateway(val, checker, []string{"http://localhost:3001"}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	defer gw.Close()

	srv := httptest.NewServer(gw)
	defer srv.Close()

	t.Run("rejects missing token with error and disconnect", func(t *testing.T) {
		conn := dialWS(t, srv.URL, nil, "")
		connectNamespace(t, conn, "")

		errPkt := readWSMsg(t, conn)
		assert.Contains(t, errPkt, `"error"`)
		assert.Contains(t, errPkt, "Authentication token required")

		discPkt := readWSMsg(t, conn)
		assert.Equal(t, "41/events", discPkt)
		assert.False(t, gw.IsUserOnline(user1.String()))
	})

	t.Run("rejects suspended account", func(t *testing.T) {
		conn := dialWS(t, srv.URL, nil, "&token=tok-susp")
		connectNamespace(t, conn, "")

		errPkt := readWSMsg(t, conn)
		assert.Contains(t, errPkt, "Account is not active")
		discPkt := readWSMsg(t, conn)
		assert.Equal(t, "41/events", discPkt)
	})

	t.Run("authenticates via cookie/auth/header and broadcasts presence", func(t *testing.T) {
		hdr1 := http.Header{}
		hdr1.Set("Cookie", "access_token=tok-u1")
		c1 := dialWS(t, srv.URL, hdr1, "")
		connectNamespace(t, c1, "")

		// Expect connected and user:online on c1
		msg1 := readWSMsg(t, c1)
		assert.Contains(t, msg1, `"connected"`)
		assert.Contains(t, msg1, user1.String())

		msg2 := readWSMsg(t, c1)
		assert.Contains(t, msg2, `"user:online"`)
		assert.Contains(t, msg2, user1.String())
		assert.True(t, gw.IsUserOnline(user1.String()))
		assert.Equal(t, 1, gw.GetConnectedUsersCount())

		// Connect second socket for user1 via auth payload -> should NOT re-emit user:online
		c1b := dialWS(t, srv.URL, nil, "")
		connectNamespace(t, c1b, `{"token":"tok-u1"}`)
		msg1b := readWSMsg(t, c1b)
		assert.Contains(t, msg1b, `"connected"`)

		// Connect user2 via Authorization header -> c1 receives user:online for user2
		hdr2 := http.Header{}
		hdr2.Set("Authorization", "Bearer tok-u2")
		c2 := dialWS(t, srv.URL, hdr2, "")
		connectNamespace(t, c2, "")
		_ = readWSMsg(t, c2) // connected
		_ = readWSMsg(t, c2) // user:online self

		onlineNotice := readWSMsg(t, c1)
		assert.Contains(t, onlineNotice, `"user:online"`)
		assert.Contains(t, onlineNotice, user2.String())
		assert.Equal(t, 2, gw.GetConnectedUsersCount())

		// Close c1 -> user1 still has c1b, so remains online
		_ = c1.Close()
		time.Sleep(50 * time.Millisecond)
		assert.True(t, gw.IsUserOnline(user1.String()))

		// Close c1b -> user1 goes offline, c2 receives user:offline
		_ = c1b.Close()
		offlineNotice := readWSMsg(t, c2)
		assert.Contains(t, offlineNotice, `"user:offline"`)
		assert.Contains(t, offlineNotice, user1.String())
		assert.False(t, gw.IsUserOnline(user1.String()))
		lastSeen, ok := gw.GetUserLastSeen(user1.String())
		assert.True(t, ok)
		assert.NotEmpty(t, lastSeen)
	})
}

func TestGateway_RoomsAndBroadcasts(t *testing.T) {
	userID := uuid.New()
	val := &stubTokenValidator{
		tokens: map[string]struct {
			sub   string
			email string
			role  string
		}{
			"tok": {sub: userID.String(), email: "u@orbito.io", role: "MEMBER"},
		},
	}

	gw := NewGateway(val, nil, []string{"*"}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	defer gw.Close()
	srv := httptest.NewServer(gw)
	defer srv.Close()

	conn := dialWS(t, srv.URL, nil, "")
	connectNamespace(t, conn, `{"token":"tok"}`)
	_ = readWSMsg(t, conn) // connected
	_ = readWSMsg(t, conn) // user:online

	projectID := uuid.New().String()
	taskID := uuid.New().String()
	wsID := uuid.New().String()
	orgID := uuid.New().String()

	// Join all 4 scopes
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`42/events,["join:organization",{"organizationId":"`+orgID+`"}]`)))
	assert.Contains(t, readWSMsg(t, conn), `"joined:organization"`)

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`42/events,["join:workspace",{"workspaceId":"`+wsID+`"}]`)))
	assert.Contains(t, readWSMsg(t, conn), `"joined:workspace"`)

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`42/events,["join:project",{"projectId":"`+projectID+`"}]`)))
	assert.Contains(t, readWSMsg(t, conn), `"joined:project"`)

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`42/events,["join:task",{"taskId":"`+taskID+`"}]`)))
	assert.Contains(t, readWSMsg(t, conn), `"joined:task"`)

	assert.Equal(t, []string{userID.String()}, gw.GetOnlineUsersInRoom("project:"+projectID))

	// Emit domain events and verify envelopes
	gw.EmitTaskCreated(projectID, map[string]any{"id": taskID, "title": "Hello"})
	pkt := readWSMsg(t, conn)
	assert.Contains(t, pkt, `"task:created"`)
	assert.Contains(t, pkt, `"Hello"`)

	gw.EmitTaskUpdated(projectID, taskID, map[string]any{"title": "Updated"})
	pktProj := readWSMsg(t, conn)
	pktTask := readWSMsg(t, conn)
	assert.Contains(t, pktProj, `"task:updated"`)
	assert.Contains(t, pktTask, `"task:updated"`)

	gw.EmitTaskDeleted(projectID, taskID)
	assert.Contains(t, readWSMsg(t, conn), `"task:deleted"`)

	gw.EmitTaskStatusChanged(projectID, taskID, map[string]any{"status": "DONE"})
	assert.Contains(t, readWSMsg(t, conn), `"task:status_changed"`)

	gw.EmitTaskAssigned(projectID, taskID, map[string]any{"assigneeId": userID.String()})
	assert.Contains(t, readWSMsg(t, conn), `"task:assigned"`)
	assert.Contains(t, readWSMsg(t, conn), `"task:assigned"`) // also emitted to user:<assigneeId>

	gw.EmitCommentAdded(projectID, taskID, map[string]any{"id": "c1"})
	assert.Contains(t, readWSMsg(t, conn), `"comment:added"`)
	assert.Contains(t, readWSMsg(t, conn), `"comment:added"`)

	gw.EmitTimeEntryStarted(projectID, taskID, map[string]any{"id": "te1"})
	assert.Contains(t, readWSMsg(t, conn), `"time:started"`)

	gw.EmitTimeEntryStopped(projectID, taskID, map[string]any{"id": "te1"})
	assert.Contains(t, readWSMsg(t, conn), `"time:stopped"`)

	gw.EmitProjectUpdated(wsID, projectID, map[string]any{"name": "P2"})
	assert.Contains(t, readWSMsg(t, conn), `"project:updated"`)

	gw.EmitSprintStarted(projectID, map[string]any{"id": "sp1"})
	assert.Contains(t, readWSMsg(t, conn), `"sprint:started"`)

	gw.EmitSprintCompleted(projectID, map[string]any{"id": "sp1"})
	assert.Contains(t, readWSMsg(t, conn), `"sprint:completed"`)

	gw.EmitNotification(userID.String(), map[string]any{"title": "Ping"})
	assert.Contains(t, readWSMsg(t, conn), `"notification"`)

	gw.EmitUserTyping(taskID, map[string]any{"id": userID.String()})
	assert.Contains(t, readWSMsg(t, conn), `"user:typing"`)

	gw.EmitUserStoppedTyping(taskID, map[string]any{"id": userID.String()})
	assert.Contains(t, readWSMsg(t, conn), `"user:stopped_typing"`)

	// Leave scopes
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`42/events,["leave:project"]`)))
	assert.Contains(t, readWSMsg(t, conn), `"left:project"`)
	assert.Empty(t, gw.GetOnlineUsersInRoom("project:"+projectID))

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`42/events,["leave:task",{"taskId":"`+taskID+`"}]`)))
	assert.Contains(t, readWSMsg(t, conn), `"left:task"`)
}

func TestGateway_PollingHandshakeAndUpgrade(t *testing.T) {
	userID := uuid.New()
	val := &stubTokenValidator{
		tokens: map[string]struct {
			sub   string
			email string
			role  string
		}{
			"poll-tok": {sub: userID.String(), email: "p@orbito.io", role: "MEMBER"},
		},
	}

	gw := NewGateway(val, nil, []string{"*"}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	defer gw.Close()
	srv := httptest.NewServer(gw)
	defer srv.Close()

	// 1. Initial polling GET handshake
	resp, err := http.Get(srv.URL + "/socket.io/?EIO=4&transport=polling")
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, strings.HasPrefix(string(body), "0{"))

	var hs eioHandshake
	require.NoError(t, json.Unmarshal(body[1:], &hs))
	require.NotEmpty(t, hs.SID)

	// 2. POST Socket.IO CONNECT on /events
	postResp, err := http.Post(
		srv.URL+"/socket.io/?EIO=4&transport=polling&sid="+hs.SID,
		"text/plain; charset=UTF-8",
		strings.NewReader(`40/events,{"token":"poll-tok"}`),
	)
	require.NoError(t, err)
	_ = postResp.Body.Close()
	require.Equal(t, http.StatusOK, postResp.StatusCode)

	// 3. GET queued packets
	pollResp, err := http.Get(srv.URL + "/socket.io/?EIO=4&transport=polling&sid=" + hs.SID)
	require.NoError(t, err)
	pollBody, err := io.ReadAll(pollResp.Body)
	_ = pollResp.Body.Close()
	require.NoError(t, err)
	assert.Contains(t, string(pollBody), "40/events,")
	assert.Contains(t, string(pollBody), `"connected"`)
	assert.Contains(t, string(pollBody), `"user:online"`)

	// 4. Upgrade polling sid to WebSocket
	wsConn := dialWS(t, srv.URL, nil, "&sid="+hs.SID)
	require.NoError(t, wsConn.WriteMessage(websocket.TextMessage, []byte("2probe")))
	assert.Equal(t, "3probe", readWSMsg(t, wsConn))
	require.NoError(t, wsConn.WriteMessage(websocket.TextMessage, []byte("5")))

	gw.EmitNotification(userID.String(), map[string]any{"title": "Upgraded"})
	assert.Contains(t, readWSMsg(t, wsConn), `"Upgraded"`)
}
