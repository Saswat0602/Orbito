package slugify_test

import (
	"testing"

	"github.com/orbito/api-go/internal/shared/slugify"
	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Acme Corporation", "acme-corporation"},
		{"  Hello   World!  ", "hello-world"},
		{"Project #1: Design & Development", "project-1-design-development"},
		{"---Special...Characters---", "special-characters"},
		{"", "item"},
		{"   ", "item"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, slugify.Slugify(tt.input))
	}
}
