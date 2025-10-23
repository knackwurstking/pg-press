package utils

import (
	"html/template"
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "simple text",
			input: "Hello world",
			contains: []string{
				"<p>Hello world</p>",
			},
		},
		{
			name:  "bold text",
			input: "This is **bold** text",
			contains: []string{
				"<strong>bold</strong>",
			},
		},
		{
			name:  "italic text",
			input: "This is *italic* text",
			contains: []string{
				"<em>italic</em>",
			},
		},
		{
			name:  "header 1",
			input: "# Main Header",
			contains: []string{
				"<h1>Main Header</h1>",
			},
		},
		{
			name:  "header 2",
			input: "## Sub Header",
			contains: []string{
				"<h2>Sub Header</h2>",
			},
		},
		{
			name:  "code inline",
			input: "Use `code` here",
			contains: []string{
				"<code>code</code>",
			},
		},
		{
			name:  "code block",
			input: "```\ncode block\n```",
			contains: []string{
				"<pre><code>code block",
			},
		},
		{
			name:  "unordered list",
			input: "- Item 1\n- Item 2",
			contains: []string{
				"<ul>",
				"<li>Item 1</li>",
				"<li>Item 2</li>",
				"</ul>",
			},
		},
		{
			name:  "ordered list",
			input: "1. First\n2. Second",
			contains: []string{
				"<ol>",
				"<li>First</li>",
				"<li>Second</li>",
				"</ol>",
			},
		},
		{
			name:  "link",
			input: "[Google](https://google.com)",
			contains: []string{
				`<a href="https://google.com">Google</a>`,
			},
		},
		{
			name:  "strikethrough",
			input: "~~strikethrough~~",
			contains: []string{
				"<del>strikethrough</del>",
			},
		},
		{
			name:  "mixed formatting",
			input: "# Header\n\nThis is **bold** and *italic* text.\n\n- List item\n- Another item",
			contains: []string{
				"<h1>Header</h1>",
				"<strong>bold</strong>",
				"<em>italic</em>",
				"<ul>",
				"<li>List item</li>",
			},
		},
		{
			name:     "empty input",
			input:    "",
			contains: []string{},
		},

		{
			name:  "double newline paragraph breaks",
			input: "Paragraph 1\n\nParagraph 2",
			contains: []string{
				"<p>Paragraph 1</p>",
				"<p>Paragraph 2</p>",
			},
		},

		{
			name:  "multiple consecutive newlines",
			input: "Line 1\n\n\nLine 2",
			contains: []string{
				"<p>Line 1</p>",
				"<p>Line 2</p>",
			},
		},
		{
			name:  "explicit line breaks in lists",
			input: "- Item 1  \nContinued on next line\n- Item 2",
			contains: []string{
				"<ul>",
				"<li>Item 1<br>",
				"Continued on next line</li>",
				"<li>Item 2</li>",
				"</ul>",
			},
		},
		{
			name:  "newlines before headers preserved",
			input: "Some text\n\n# Header\n\nMore text",
			contains: []string{
				"<p>Some text</p>",
				"<h1>Header</h1>",
				"<p>More text</p>",
			},
		},
		{
			name:  "newlines_before_lists_preserved",
			input: "Some text\n\n- List item\n- Another item",
			contains: []string{
				"<p>Some text</p>",
				"<ul>",
				"<li>List item</li>",
				"<li>Another item</li>",
				"</ul>",
			},
		},

		{
			name:  "heading followed by list without empty line",
			input: "### Issues Found\n- First issue\n- Second issue",
			contains: []string{
				"<h3>Issues Found</h3>",
				"<ul>",
				"<li>First issue</li>",
				"<li>Second issue</li>",
				"</ul>",
			},
		},
		{
			name:  "h1 heading followed by ordered list",
			input: "# Main Issues\n1. First problem\n2. Second problem",
			contains: []string{
				"<h1>Main Issues</h1>",
				"<ol>",
				"<li>First problem</li>",
				"<li>Second problem</li>",
				"</ol>",
			},
		},
		{
			name:  "h2 heading followed by unordered list with asterisks",
			input: "## Problems\n* Item one\n* Item two",
			contains: []string{
				"<h2>Problems</h2>",
				"<ul>",
				"<li>Item one</li>",
				"<li>Item two</li>",
				"</ul>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderMarkdown(tt.input)

			// Check that result is of type template.HTML
			if _, ok := interface{}(result).(template.HTML); !ok {
				t.Errorf("RenderMarkdown() should return template.HTML, got %T", result)
			}

			resultStr := string(result)

			// Check that all expected strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("RenderMarkdown() result should contain %q, got: %s", expected, resultStr)
				}
			}
		})
	}
}

func TestRenderMarkdownReturnsHTML(t *testing.T) {
	input := "**bold**"
	result := RenderMarkdown(input)

	// Verify it returns template.HTML type
	_, ok := interface{}(result).(template.HTML)
	if !ok {
		t.Errorf("RenderMarkdown should return template.HTML, got %T", result)
	}

	// Verify HTML content
	if !strings.Contains(string(result), "<strong>bold</strong>") {
		t.Errorf("Expected HTML output, got: %s", result)
	}
}

func TestRenderMarkdownWithSpecialCharacters(t *testing.T) {
	input := "Test with <script>alert('xss')</script> and & symbols"
	result := RenderMarkdown(input)

	resultStr := string(result)

	// The markdown renderer should handle special characters safely
	// We expect the HTML to be properly escaped or handled
	if strings.Contains(resultStr, "<script>") {
		t.Errorf("Potential XSS vulnerability: script tags should be escaped or sanitized")
	}
}

func BenchmarkRenderMarkdown(b *testing.B) {
	input := "# Header\n\nThis is a **benchmark test** with *italic* text.\n\n## Subheader\n\n- List item 1\n- List item 2\n- List item 3\n\nSome code: `fmt.Println(\"Hello\")`\n\n```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderMarkdown(input)
	}
}
