package commentutil

import (
	"fmt"
	"log"
	"strings"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/proto/rdf"
)

// `path` to `position`(Lnum for new file) to comment `body`s
type PostedComments map[string]map[int][]string

// IsPosted returns true if a given comment has been posted in code review service already,
// otherwise returns false. It sees comments with same path, same position,
// and same body as same comments.
func (p PostedComments) IsPosted(c *reviewdog.Comment, lineNum int, body string) bool {
	path := c.Result.Diagnostic.GetLocation().GetPath()
	if _, ok := p[path]; !ok {
		return false
	}
	bodies, ok := p[path][lineNum]
	if !ok {
		return false
	}
	for _, b := range bodies {
		if b == body {
			return true
		}
	}
	return false
}

// AddPostedComment adds a posted comment.
func (p PostedComments) AddPostedComment(path string, lineNum int, body string) {
	if _, ok := p[path]; !ok {
		p[path] = make(map[int][]string)
	}
	if _, ok := p[path][lineNum]; !ok {
		p[path][lineNum] = make([]string, 0)
	}
	p[path][lineNum] = append(p[path][lineNum], body)
}

// DebugLog outputs posted comments as log for debugging.
func (p PostedComments) DebugLog() {
	for filename, f := range p {
		for line := range f {
			log.Printf("[debug] posted: %s:%d", filename, line)
		}
	}
}

// BodyPrefix is prefix text of comment body.
const BodyPrefix = `<sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>`

// MarkdownComment creates comment body markdown.
func MarkdownComment(c *reviewdog.Comment) string {
	var sb strings.Builder
	if s := severity(c); s != "" {
		sb.WriteString(s)
		sb.WriteString(" ")
	}
	if tool := toolName(c); tool != "" {
		sb.WriteString(fmt.Sprintf("**[%s]** ", tool))
	}
	if code := c.Result.Diagnostic.GetCode().GetValue(); code != "" {
		if url := c.Result.Diagnostic.GetCode().GetUrl(); url != "" {
			sb.WriteString(fmt.Sprintf("<[%s](%s)> ", code, url))
		} else {
			sb.WriteString(fmt.Sprintf("<%s> ", code))
		}
	}
	sb.WriteString(BodyPrefix)
	sb.WriteString(c.Result.Diagnostic.GetMessage())
	return sb.String()
}

// BitBucketMarkdownComment creates BitBucket markdown comment (it has very restricted markdown support in comments)
func BitBucketMarkdownComment(c *reviewdog.Comment) string {
	var sb strings.Builder
	if s := severity(c); s != "" {
		sb.WriteString(s)
		sb.WriteString(" ")
	}
	if tool := toolName(c); tool != "" {
		sb.WriteString(fmt.Sprintf("**[%s]** ", tool))
	}
	if code := c.Result.Diagnostic.GetCode().GetValue(); code != "" {
		if url := c.Result.Diagnostic.GetCode().GetUrl(); url != "" {
			sb.WriteString(fmt.Sprintf("<[%s](%s)> ", code, url))
		} else {
			sb.WriteString(fmt.Sprintf("<%s> ", code))
		}
	}
	// Add markdown's hard line break
	sb.WriteString("  \n")
	sb.WriteString(c.Result.Diagnostic.GetMessage())
	return sb.String()
}

func toolName(c *reviewdog.Comment) string {
	if name := c.Result.Diagnostic.GetSource().GetName(); name != "" {
		return name
	}
	return c.ToolName
}

func severity(c *reviewdog.Comment) string {
	switch c.Result.Diagnostic.GetSeverity() {
	case rdf.Severity_ERROR:
		return "🚫"
	case rdf.Severity_WARNING:
		return "⚠️"
	case rdf.Severity_INFO:
		return "📝"
	default:
		return ""
	}
}
