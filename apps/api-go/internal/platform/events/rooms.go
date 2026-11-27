package events

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/orbito/api-go/internal/shared/httpx"
)

type roomTarget struct {
	OrganizationID string `json:"organizationId"`
	WorkspaceID    string `json:"workspaceId"`
	ProjectID      string `json:"projectId"`
	TaskID         string `json:"taskId"`
}

func nowISO() string {
	return time.Now().UTC().Format(httpx.ISO8601Format)
}

func (g *Gateway) registerAuthenticatedClient(sess *eioSession, userID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sess.userID = userID
	userSockets, exists := g.connectedUsers[userID]
	if !exists {
		userSockets = make(map[string]struct{})
		g.connectedUsers[userID] = userSockets
	}
	wasOffline := len(userSockets) == 0
	userSockets[sess.nsSID] = struct{}{}
	g.clients[sess.nsSID] = sess

	personalRoom := "user:" + userID
	g.addToRoomLocked(personalRoom, sess)
	return wasOffline
}

func (g *Gateway) unregisterClient(sess *eioSession) {
	if sess == nil {
		return
	}

	var leftUserID string
	var wentOffline bool
	var lastSeen string

	g.mu.Lock()
	userID := sess.userID
	nsSID := sess.nsSID
	if nsSID != "" {
		delete(g.clients, nsSID)
		for room := range sess.rooms {
			g.removeFromRoomLocked(room, nsSID)
		}
		sess.rooms = make(map[string]struct{})
	}
	if userID != "" && nsSID != "" {
		if sockets, ok := g.connectedUsers[userID]; ok {
			delete(sockets, nsSID)
			if len(sockets) == 0 {
				delete(g.connectedUsers, userID)
				lastSeen = nowISO()
				g.userLastSeen[userID] = lastSeen
				leftUserID = userID
				wentOffline = true
			}
		}
	}
	sess.userID = ""
	sess.nsSID = ""
	g.mu.Unlock()

	if wentOffline {
		g.broadcastAll("user:offline", map[string]any{
			"userId":   leftUserID,
			"isOnline": false,
			"lastSeen": lastSeen,
		})
	}
}

func (g *Gateway) addToRoomLocked(room string, sess *eioSession) {
	if room == "" || sess.nsSID == "" {
		return
	}
	members, ok := g.rooms[room]
	if !ok {
		members = make(map[string]*eioSession)
		g.rooms[room] = members
	}
	members[sess.nsSID] = sess
	if sess.rooms == nil {
		sess.rooms = make(map[string]struct{})
	}
	sess.rooms[room] = struct{}{}
}

func (g *Gateway) removeFromRoomLocked(room, nsSID string) {
	members, ok := g.rooms[room]
	if !ok {
		return
	}
	delete(members, nsSID)
	if len(members) == 0 {
		delete(g.rooms, room)
	}
}

func (g *Gateway) joinRoom(sess *eioSession, room string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.addToRoomLocked(room, sess)
}

func (g *Gateway) leaveRoom(sess *eioSession, room string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.removeFromRoomLocked(room, sess.nsSID)
	delete(sess.rooms, room)
}

func (g *Gateway) handleClientEvent(sess *eioSession, eventName string, rawPayload json.RawMessage) {
	if sess.userID == "" {
		g.emitToSession(sess, "error", map[string]any{"message": "Authentication required"})
		return
	}

	var target roomTarget
	if len(rawPayload) > 0 && string(rawPayload) != "null" {
		if err := json.Unmarshal(rawPayload, &target); err != nil {
			if g.log != nil {
				g.log.Debug("failed to decode socket event payload", slog.String("event", eventName), slog.String("error", err.Error()))
			}
		}
	}

	switch eventName {
	case "join:organization":
		if target.OrganizationID == "" {
			g.emitToSession(sess, "error", map[string]any{"message": "Organization ID is required"})
			return
		}
		g.mu.Lock()
		sess.organizationID = target.OrganizationID
		g.addToRoomLocked("org:"+target.OrganizationID, sess)
		g.mu.Unlock()
		g.emitToSession(sess, "joined:organization", map[string]any{"organizationId": target.OrganizationID})

	case "leave:organization":
		g.mu.Lock()
		orgID := sess.organizationID
		if orgID == "" {
			g.mu.Unlock()
			g.emitToSession(sess, "error", map[string]any{"message": "Not joined to any organization"})
			return
		}
		g.removeFromRoomLocked("org:"+orgID, sess.nsSID)
		delete(sess.rooms, "org:"+orgID)
		sess.organizationID = ""
		g.mu.Unlock()
		g.emitToSession(sess, "left:organization", map[string]any{"organizationId": orgID})

	case "join:workspace":
		if target.WorkspaceID == "" {
			g.emitToSession(sess, "error", map[string]any{"message": "Workspace ID is required"})
			return
		}
		g.mu.Lock()
		sess.workspaceID = target.WorkspaceID
		g.addToRoomLocked("workspace:"+target.WorkspaceID, sess)
		g.mu.Unlock()
		g.emitToSession(sess, "joined:workspace", map[string]any{"workspaceId": target.WorkspaceID})

	case "leave:workspace":
		g.mu.Lock()
		wsID := sess.workspaceID
		if wsID == "" {
			g.mu.Unlock()
			g.emitToSession(sess, "error", map[string]any{"message": "Not joined to any workspace"})
			return
		}
		g.removeFromRoomLocked("workspace:"+wsID, sess.nsSID)
		delete(sess.rooms, "workspace:"+wsID)
		sess.workspaceID = ""
		g.mu.Unlock()
		g.emitToSession(sess, "left:workspace", map[string]any{"workspaceId": wsID})

	case "join:project":
		if target.ProjectID == "" {
			g.emitToSession(sess, "error", map[string]any{"message": "Project ID is required"})
			return
		}
		g.mu.Lock()
		sess.projectID = target.ProjectID
		g.addToRoomLocked("project:"+target.ProjectID, sess)
		g.mu.Unlock()
		g.emitToSession(sess, "joined:project", map[string]any{"projectId": target.ProjectID})

	case "leave:project":
		g.mu.Lock()
		projID := sess.projectID
		if projID == "" {
			g.mu.Unlock()
			g.emitToSession(sess, "error", map[string]any{"message": "Not joined to any project"})
			return
		}
		g.removeFromRoomLocked("project:"+projID, sess.nsSID)
		delete(sess.rooms, "project:"+projID)
		sess.projectID = ""
		g.mu.Unlock()
		g.emitToSession(sess, "left:project", map[string]any{"projectId": projID})

	case "join:task":
		if target.TaskID == "" {
			g.emitToSession(sess, "error", map[string]any{"message": "Task ID is required"})
			return
		}
		g.joinRoom(sess, "task:"+target.TaskID)
		g.emitToSession(sess, "joined:task", map[string]any{"taskId": target.TaskID})

	case "leave:task":
		if target.TaskID == "" {
			g.emitToSession(sess, "error", map[string]any{"message": "Task ID is required"})
			return
		}
		g.leaveRoom(sess, "task:"+target.TaskID)
		g.emitToSession(sess, "left:task", map[string]any{"taskId": target.TaskID})

	case "user:typing", "user:stopped_typing":
		if target.TaskID == "" {
			return
		}
		var rawMap map[string]any
		_ = json.Unmarshal(rawPayload, &rawMap)
		if eventName == "user:typing" {
			g.EmitUserTyping(target.TaskID, rawMap["user"])
		} else {
			g.EmitUserStoppedTyping(target.TaskID, rawMap["user"])
		}
	}
}

// GetConnectedUsersCount returns the number of distinct users with active sockets.
func (g *Gateway) GetConnectedUsersCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.connectedUsers)
}

// IsUserOnline returns true if the given user currently has at least one active socket.
func (g *Gateway) IsUserOnline(userID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	sockets, ok := g.connectedUsers[userID]
	return ok && len(sockets) > 0
}

// GetUserLastSeen returns the last recorded disconnect timestamp for a user, if any.
func (g *Gateway) GetUserLastSeen(userID string) (string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	ts, ok := g.userLastSeen[userID]
	return ts, ok
}

// GetOnlineUsersInRoom returns distinct authenticated user IDs currently joined to a room.
func (g *Gateway) GetOnlineUsersInRoom(room string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	members, ok := g.rooms[room]
	if !ok {
		return []string{}
	}
	seen := make(map[string]struct{}, len(members))
	out := make([]string, 0, len(members))
	for _, sess := range members {
		if sess.userID != "" {
			if _, dup := seen[sess.userID]; !dup {
				seen[sess.userID] = struct{}{}
				out = append(out, sess.userID)
			}
		}
	}
	return out
}

