package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	log := New(true, &buf)

	log.Info("test message", "key", "value")

	var data map[string]any
	err := json.Unmarshal(buf.Bytes(), &data)
	require.NoError(t, err)

	assert.Equal(t, "INFO", data["level"])
	assert.Equal(t, "test message", data["msg"])
	assert.Equal(t, "value", data["key"])
}

func TestContextLogger(t *testing.T) {
	var buf bytes.Buffer
	log := New(false, &buf)

	ctx := WithContext(context.Background(), log)
	extracted := FromContext(ctx)
	assert.Equal(t, log, extracted)

	// Test fallback
	fallback := FromContext(context.Background())
	assert.NotNil(t, fallback)
}
