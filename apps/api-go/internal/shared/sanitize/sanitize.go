package sanitize

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

var allowedTags = map[string]bool{
	"p": true, "div": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"blockquote": true, "pre": true, "code": true, "ul": true, "ol": true, "li": true,
	"table": true, "thead": true, "tbody": true, "tfoot": true, "tr": true, "th": true, "td": true, "hr": true,
	"a": true, "b": true, "strong": true, "i": true, "em": true, "u": true, "s": true,
	"del": true, "ins": true, "span": true, "br": true, "sub": true, "sup": true, "mark": true, "kbd": true, "strike": true,
	"img": true, "input": true,
}

var allowedGlobalAttrs = map[string]bool{
	"class": true, "style": true, "title": true, "classname": true,
}

var allowedTagAttrs = map[string]map[string]bool{
	"a":     {"href": true, "name": true, "target": true, "title": true, "rel": true, "data-id": true, "data-type": true},
	"img":   {"src": true, "alt": true, "width": true, "height": true},
	"input": {"type": true, "checked": true, "disabled": true},
	"code":  {"classname": true},
}

func isSafeURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if lower == "" || strings.HasPrefix(lower, "/") || strings.HasPrefix(lower, "#") {
		return true
	}
	for _, scheme := range []string{"http:", "https:", "mailto:", "data:"} {
		if strings.HasPrefix(lower, scheme) {
			return true
		}
	}
	return false
}

// HTML keeps allowed rich-text tags/attributes and strips dangerous elements like script/style/iframe.
func HTML(input string) string {
	if input == "" {
		return input
	}
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var out bytes.Buffer
	skipDepth := 0

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return out.String()
		case html.TextToken:
			if skipDepth == 0 {
				out.WriteString(html.EscapeString(string(tokenizer.Text())))
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := tokenizer.Token()
			tag := strings.ToLower(tok.Data)
			if tag == "script" || tag == "style" || tag == "iframe" {
				if tt == html.StartTagToken {
					skipDepth++
				}
				continue
			}
			if skipDepth > 0 || !allowedTags[tag] {
				continue
			}
			var cleanAttrs []html.Attribute
			tagMap := allowedTagAttrs[tag]
			for _, attr := range tok.Attr {
				key := strings.ToLower(attr.Key)
				if !allowedGlobalAttrs[key] && (tagMap == nil || !tagMap[key]) {
					continue
				}
				if (key == "href" || key == "src") && !isSafeURL(attr.Val) {
					continue
				}
				cleanAttrs = append(cleanAttrs, attr)
			}
			tok.Attr = cleanAttrs
			out.WriteString(tok.String())
		case html.EndTagToken:
			tok := tokenizer.Token()
			tag := strings.ToLower(tok.Data)
			if tag == "script" || tag == "style" || tag == "iframe" {
				if skipDepth > 0 {
					skipDepth--
				}
				continue
			}
			if skipDepth == 0 && allowedTags[tag] {
				out.WriteString("</" + tag + ">")
			}
		}
	}
}

// Text strips all HTML tags from plain-text fields such as titles.
func Text(input string) string {
	if input == "" {
		return input
	}
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var out strings.Builder
	skipDepth := 0
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return out.String()
		case html.StartTagToken:
			tag := strings.ToLower(tokenizer.Token().Data)
			if tag == "script" || tag == "style" {
				skipDepth++
			}
		case html.EndTagToken:
			tag := strings.ToLower(tokenizer.Token().Data)
			if (tag == "script" || tag == "style") && skipDepth > 0 {
				skipDepth--
			}
		case html.TextToken:
			if skipDepth == 0 {
				out.Write(tokenizer.Text())
			}
		}
	}
}

// Object recursively sanitizes all string values inside a JSON-compatible map or slice.
func Object(val any) any {
	switch v := val.(type) {
	case string:
		return Text(v)
	case []any:
		res := make([]any, len(v))
		for i, item := range v {
			res[i] = Object(item)
		}
		return res
	case map[string]any:
		res := make(map[string]any, len(v))
		for k, item := range v {
			res[k] = Object(item)
		}
		return res
	default:
		return val
	}
}
