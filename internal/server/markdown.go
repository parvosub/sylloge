package server

import (
	"html"
	"strings"
)

// summaryToHTML converts the markdown produced by the summarizer into a
// flowing, prose-style HTML representation for the editable summary box.
// It is intentionally specific to the model's output style:
//   - headings (**Heading** or ### Heading) become <h4> elements styled as inline bold text
//   - multi-line paragraphs under a heading are merged into one paragraph
//   - * bullet lists under a heading are flattened into a single flowing paragraph
//   - inline **bold** text is preserved as <strong>
func summaryToHTML(md string) string {
	md = strings.TrimSpace(md)
	if md == "" {
		return ""
	}

	lines := strings.Split(md, "\n")
	var blocks []string
	var currentBlock []string

	flushBlock := func() {
		if len(currentBlock) == 0 {
			return
		}

		firstLine := stripBullet(strings.TrimSpace(currentBlock[0]))
		if headingText, ok := extractHeading(firstLine); ok && len(currentBlock) > 1 {
			blocks = append(blocks, headingHTML(headingText))
			rest := currentBlock[1:]

			if allBullets(rest) {
				blocks = append(blocks, bulletListToParagraph(rest))
			} else {
				// Merge consecutive non-heading lines into paragraphs.
				var paraLines []string
				flushPara := func() {
					if len(paraLines) == 0 {
						return
					}
					content := strings.Join(paraLines, " ")
					content = strings.TrimSpace(content)
					if content != "" {
						blocks = append(blocks, "<p>"+applyBold(content)+"</p>")
					}
					paraLines = nil
				}
				for _, line := range rest {
					line = stripBullet(strings.TrimSpace(line))
					if line == "" {
						flushPara()
						continue
					}
					if h, ok := extractHeading(line); ok {
						flushPara()
						blocks = append(blocks, headingHTML(h))
						continue
					}
					paraLines = append(paraLines, line)
				}
				flushPara()
			}
			currentBlock = nil
			return
		}

		// No leading heading: process each line independently.
		for _, line := range currentBlock {
			line = stripBullet(line)
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if headingText, ok := extractHeading(line); ok {
				blocks = append(blocks, headingHTML(headingText))
				continue
			}
			blocks = append(blocks, "<p>"+applyBold(line)+"</p>")
		}
		currentBlock = nil
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flushBlock()
		} else {
			currentBlock = append(currentBlock, line)
		}
	}
	flushBlock()

	return strings.Join(blocks, "\n")
}

func stripBullet(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "*   ") {
		return strings.TrimSpace(strings.TrimPrefix(s, "*   "))
	}
	if strings.HasPrefix(s, "* ") {
		return strings.TrimSpace(strings.TrimPrefix(s, "* "))
	}
	return s
}

func isBullet(s string) bool {
	return strings.HasPrefix(strings.TrimSpace(s), "* ") || strings.HasPrefix(strings.TrimSpace(s), "*   ")
}

func allBullets(lines []string) bool {
	hasBullet := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !isBullet(trimmed) {
			return false
		}
		hasBullet = true
	}
	return hasBullet
}

func bulletListToParagraph(lines []string) string {
	var items []string
	for _, line := range lines {
		line = stripBullet(line)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasSuffix(line, ".") && !strings.HasSuffix(line, "!") && !strings.HasSuffix(line, "?") {
			line += "."
		}
		items = append(items, applyBold(line))
	}
	if len(items) == 0 {
		return ""
	}
	return "<p>" + strings.Join(items, " ") + "</p>"
}

func extractHeading(s string) (string, bool) {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "**") && strings.HasSuffix(s, "**") {
		inner := s[2 : len(s)-2]
		if inner != "" && !strings.Contains(inner, "**") {
			return inner, true
		}
	}

	for _, prefix := range []string{"### ", "## ", "# "} {
		if strings.HasPrefix(s, prefix) {
			inner := strings.TrimSpace(strings.TrimPrefix(s, prefix))
			if inner != "" {
				return inner, true
			}
		}
	}
	return "", false
}

func headingHTML(text string) string {
	return "<h4 class=\"summary-heading\">" + html.EscapeString(text) + "</h4>"
}

func applyBold(s string) string {
	var result strings.Builder
	for {
		start := strings.Index(s, "**")
		if start == -1 {
			result.WriteString(html.EscapeString(s))
			break
		}
		result.WriteString(html.EscapeString(s[:start]))
		s = s[start+2:]
		end := strings.Index(s, "**")
		if end == -1 {
			result.WriteString("**")
			result.WriteString(html.EscapeString(s))
			break
		}
		result.WriteString("<strong>")
		result.WriteString(html.EscapeString(s[:end]))
		result.WriteString("</strong>")
		s = s[end+2:]
	}
	return result.String()
}
