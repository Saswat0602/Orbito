package ctxhelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, GetUserID(ctx))
	assert.Empty(t, GetUserEmail(ctx))
	assert.Empty(t, GetUserRole(ctx))
	assert.Empty(t, GetCorrelationID(ctx))

	ctx = WithUserID(ctx, "usr-12345")
	ctx = WithUserEmail(ctx, "admin@orbito.io")
	ctx = WithUserRole(ctx, "SUPER_ADMIN")

	assert.Equal(t, "usr-12345", GetUserID(ctx))
	assert.Equal(t, "admin@orbito.io", GetUserEmail(ctx))
	assert.Equal(t, "SUPER_ADMIN", GetUserRole(ctx))
}
