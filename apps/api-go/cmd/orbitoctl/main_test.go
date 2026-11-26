package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrbitoctlVersion(t *testing.T) {
	assert.Equal(t, "1.0.0", version)
}

func TestPrintUsage(t *testing.T) {
	assert.NotPanics(t, func() {
		printUsage()
	})
}
