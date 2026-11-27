package storage

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider_SaveStreamDeleteAndTraversal(t *testing.T) {
	dir := t.TempDir()
	p, err := NewLocalProvider(dir)
	require.NoError(t, err)

	res, err := p.SaveFile("uploads/tasks/abc", "note.txt", []byte("hello orbito"))
	require.NoError(t, err)
	assert.Equal(t, "uploads/tasks/abc/note.txt", res.Key)
	require.NotNil(t, res.URL)
	assert.Equal(t, "/uploads/tasks/abc/note.txt", *res.URL)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/download", nil)
	err = p.StreamLocal(rr, req, res.Key, false)
	require.NoError(t, err)
	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Header().Get("Content-Disposition"), `inline; filename="note.txt"`)
	assert.Equal(t, "hello orbito", rr.Body.String())

	// HTML must never be served inline
	_, err = p.SaveFile("uploads/tasks/abc", "xss.html", []byte("<script>alert(1)</script>"))
	require.NoError(t, err)
	rr2 := httptest.NewRecorder()
	err = p.StreamLocal(rr2, req, "uploads/tasks/abc/xss.html", false)
	require.NoError(t, err)
	assert.Equal(t, "application/octet-stream", rr2.Header().Get("Content-Type"))
	assert.Contains(t, rr2.Header().Get("Content-Disposition"), "attachment;")

	// Path traversal must fail
	_, err = p.SaveFile("../escape", "evil.txt", []byte("x"))
	require.Error(t, err)
	_, err = p.SaveFile("uploads", "../evil.txt", []byte("x"))
	require.Error(t, err)
	err = p.StreamLocal(httptest.NewRecorder(), req, "../../etc/passwd", true)
	require.Error(t, err)

	require.NoError(t, p.DeleteFile(res.Key))
	err = p.StreamLocal(httptest.NewRecorder(), req, res.Key, true)
	require.Error(t, err)
}

