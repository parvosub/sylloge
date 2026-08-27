package server

import (
	"strings"
	"testing"
)

func TestSummaryToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		omits    []string
	}{
		{
			name: "headings become h4 with following paragraph",
			input: "**Student Performance Summary**\n\n**Overview**\nAlice excels in watercolor.",
			contains: []string{
				"<h4 class=\"summary-heading\">Student Performance Summary</h4>",
				"<h4 class=\"summary-heading\">Overview</h4>",
				"<p>Alice excels in watercolor.</p>",
			},
			omits: []string{"**"},
		},
		{
			name: "bullet lists are flattened to flowing paragraph",
			input: "**Key Strengths**\n*   **Technical:** Excels in color.\n*   **Theory:** Strong grasp.",
			contains: []string{
				"<h4 class=\"summary-heading\">Key Strengths</h4>",
				"<p><strong>Technical:</strong> Excels in color. <strong>Theory:</strong> Strong grasp.</p>",
			},
			omits: []string{"<ul>", "<li>", "**", "*   "},
		},
		{
			name: "bullet items without punctuation get a period",
			input: "**Key Strengths**\n*   **Technical:** Excels in color\n*   **Theory:** Strong grasp",
			contains: []string{
				"<p><strong>Technical:</strong> Excels in color. <strong>Theory:</strong> Strong grasp.</p>",
			},
			omits: []string{"**"},
		},
		{
			name: "multi-line paragraph under heading is merged",
			input: "**Overview**\nAlice excels in\nwatercolor and theory.",
			contains: []string{
				"<h4 class=\"summary-heading\">Overview</h4>",
				"<p>Alice excels in watercolor and theory.</p>",
			},
		},
		{
			name: "bold inline text stays bold",
			input: "Alice is **strong** in watercolor and **growing** in theory.",
			contains: []string{
				"<p>Alice is <strong>strong</strong> in watercolor and <strong>growing</strong> in theory.</p>",
			},
			omits: []string{"**"},
		},
		{
			name: "empty input returns empty",
			input: "   ",
			contains: []string{},
			omits:    []string{"<p>"},
		},
		{
			name: "plain text becomes paragraph",
			input: "A fine summary.",
			contains: []string{"<p>A fine summary.</p>"},
			omits:    []string{"**"},
		},
		{
			name: "markdown heading levels become h4",
			input: "### Student Performance Summary: Alice\n\nAlice is great.",
			contains: []string{
				"<h4 class=\"summary-heading\">Student Performance Summary: Alice</h4>",
				"<p>Alice is great.</p>",
			},
			omits: []string{"###", "**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summaryToHTML(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, got)
				}
			}
			for _, omit := range tt.omits {
				if strings.Contains(got, omit) {
					t.Errorf("expected output to omit %q, got:\n%s", omit, got)
				}
			}
		})
	}
}
