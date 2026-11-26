package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizers(t *testing.T) {
	htmlRes := HTML(`<p onclick="evil()">Hello</p><script>alert(1)</script><a href="javascript:alert(2)">bad</a><a href="https://example.com">ok</a>`)
	assert.Contains(t, htmlRes, "<p>Hello</p>")
	assert.NotContains(t, htmlRes, "<script>")
	assert.NotContains(t, htmlRes, "onclick")
	assert.NotContains(t, htmlRes, "javascript:")
	assert.Contains(t, htmlRes, `href="https://example.com"`)

	assert.Equal(t, "Hello World", Text("<b>Hello</b> <script>x</script>World"))

	obj := Object(map[string]any{
		"env":  "<b>prod</b>",
		"tags": []any{"<i>a</i>", 123},
	}).(map[string]any)
	assert.Equal(t, "prod", obj["env"])
	assert.Equal(t, []any{"a", 123}, obj["tags"])
}
