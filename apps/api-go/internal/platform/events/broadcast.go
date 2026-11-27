package events

import (
	"encoding/json"
	"log/slog"
)

func (g *Gateway) encodeSocketEvent(eventName string, payload any) (string, bool) {
	raw, err := json.Marshal([]any{eventName, payload})
	if err != nil {
		if g.log != nil {
			g.log.Error("failed to marshal socket event", slog.String("event", eventName), slog.String("error", err.Error()))
		}
		return "", false
	}
	return "42" + namespaceEvents + "," + string(raw), true
}

func (g *Gateway) emitToSession(sess *eioSession, eventName string, payload any) {
	pkt, ok := g.encodeSocketEvent(eventName, payload)
	if !ok {
		return
	}
	sess.enqueue(pkt)
}

func (g *Gateway) emitToRoom(room, eventName string, data any) {
	pkt, ok := g.encodeSocketEvent(eventName, map[string]any{
		"event":     eventName,
		"data":      data,
		"timestamp": nowISO(),
	})
	if !ok {
		return
	}

	g.mu.RLock()
	members := g.rooms[room]
	targets := make([]*eioSession, 0, len(members))
	for _, sess := range members {
		targets = append(targets, sess)
	}
	g.mu.RUnlock()

	for _, sess := range targets {
		sess.enqueue(pkt)
	}
}

func (g *Gateway) broadcastAll(eventName string, data any) {
	pkt, ok := g.encodeSocketEvent(eventName, data)
	if !ok {
		return
	}

	g.mu.RLock()
	targets := make([]*eioSession, 0, len(g.clients))
	for _, sess := range g.clients {
		targets = append(targets, sess)
	}
	g.mu.RUnlock()

	for _, sess := range targets {
		sess.enqueue(pkt)
	}
}

// EmitTaskCreated broadcasts task:created to the project room.
func (g *Gateway) EmitTaskCreated(projectID string, task any) {
	g.emitToRoom("project:"+projectID, "task:created", task)
}

// EmitTaskUpdated broadcasts task:updated to both project and task rooms.
func (g *Gateway) EmitTaskUpdated(projectID, taskID string, updates any) {
	payload := map[string]any{"taskId": taskID, "updates": updates}
	g.emitToRoom("project:"+projectID, "task:updated", payload)
	g.emitToRoom("task:"+taskID, "task:updated", payload)
}

// EmitTaskDeleted broadcasts task:deleted to the project room.
func (g *Gateway) EmitTaskDeleted(projectID, taskID string) {
	g.emitToRoom("project:"+projectID, "task:deleted", map[string]any{"taskId": taskID})
}

// EmitTaskStatusChanged broadcasts task:status_changed to the project room.
func (g *Gateway) EmitTaskStatusChanged(projectID, taskID string, statusChange any) {
	g.emitToRoom("project:"+projectID, "task:status_changed", map[string]any{
		"taskId":       taskID,
		"statusChange": statusChange,
	})
}

// EmitTaskAssigned broadcasts task:assigned to the project room and directly to the assignee if present.
func (g *Gateway) EmitTaskAssigned(projectID, taskID string, assignment any) {
	payload := map[string]any{"taskId": taskID, "assignment": assignment}
	g.emitToRoom("project:"+projectID, "task:assigned", payload)
	if m, ok := assignment.(map[string]any); ok {
		if assigneeID, okStr := m["assigneeId"].(string); okStr && assigneeID != "" {
			g.emitToRoom("user:"+assigneeID, "task:assigned", payload)
		}
	}
}

// EmitCommentAdded broadcasts comment:added to both project and task rooms.
func (g *Gateway) EmitCommentAdded(projectID, taskID string, comment any) {
	payload := map[string]any{"taskId": taskID, "comment": comment}
	g.emitToRoom("project:"+projectID, "comment:added", payload)
	g.emitToRoom("task:"+taskID, "comment:added", payload)
}

// EmitTimeEntryStarted broadcasts time:started to the project room.
func (g *Gateway) EmitTimeEntryStarted(projectID, taskID string, timeEntry any) {
	g.emitToRoom("project:"+projectID, "time:started", map[string]any{
		"taskId":    taskID,
		"timeEntry": timeEntry,
	})
}

// EmitTimeEntryStopped broadcasts time:stopped to the project room.
func (g *Gateway) EmitTimeEntryStopped(projectID, taskID string, timeEntry any) {
	g.emitToRoom("project:"+projectID, "time:stopped", map[string]any{
		"taskId":    taskID,
		"timeEntry": timeEntry,
	})
}

// EmitProjectUpdated broadcasts project:updated to the workspace room.
func (g *Gateway) EmitProjectUpdated(workspaceID, projectID string, updates any) {
	g.emitToRoom("workspace:"+workspaceID, "project:updated", map[string]any{
		"projectId": projectID,
		"updates":   updates,
	})
}

// EmitSprintStarted broadcasts sprint:started to the project room.
func (g *Gateway) EmitSprintStarted(projectID string, sprint any) {
	g.emitToRoom("project:"+projectID, "sprint:started", sprint)
}

// EmitSprintCompleted broadcasts sprint:completed to the project room.
func (g *Gateway) EmitSprintCompleted(projectID string, sprint any) {
	g.emitToRoom("project:"+projectID, "sprint:completed", sprint)
}

// EmitNotification broadcasts a real-time notification to the target user's personal room.
func (g *Gateway) EmitNotification(userID string, notification any) {
	g.emitToRoom("user:"+userID, "notification", notification)
}

// EmitUserTyping broadcasts user:typing to the task room.
func (g *Gateway) EmitUserTyping(taskID string, user any) {
	g.emitToRoom("task:"+taskID, "user:typing", map[string]any{
		"taskId": taskID,
		"user":   user,
	})
}

// EmitUserStoppedTyping broadcasts user:stopped_typing to the task room.
func (g *Gateway) EmitUserStoppedTyping(taskID string, user any) {
	g.emitToRoom("task:"+taskID, "user:stopped_typing", map[string]any{
		"taskId": taskID,
		"user":   user,
	})
}

