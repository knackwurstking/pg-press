package utils

import (
	"html/template"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// RenderMarkdown converts markdown text to HTML with safe rendering
func RenderMarkdown(markdown string) template.HTML {
	// Preprocess to fix heading/list spacing
	processedMarkdown := preprocessMarkdown(markdown)

	// Configure blackfriday with safe settings
	extensions := blackfriday.NoIntraEmphasis |
		blackfriday.Tables |
		blackfriday.FencedCode |
		blackfriday.Autolink |
		blackfriday.Strikethrough |
		blackfriday.SpaceHeadings |
		blackfriday.BackslashLineBreak

	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.HTMLFlagsNone,
	})

	// Process the markdown
	htmlBytes := blackfriday.Run([]byte(processedMarkdown), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(extensions))
	htmlContent := string(htmlBytes)

	// Apply additional sanitization to the generated HTML
	sanitizedHTML := sanitizeHTML(htmlContent)

	return template.HTML(sanitizedHTML)
}

// sanitizeHTML removes potentially dangerous HTML tags and attributes
func sanitizeHTML(htmlContent string) string {
	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	htmlContent = scriptRegex.ReplaceAllString(htmlContent, "")

	// Remove standalone script tags
	scriptTagRegex := regexp.MustCompile(`(?i)</?script[^>]*>`)
	htmlContent = scriptTagRegex.ReplaceAllString(htmlContent, "")

	// Remove other potentially dangerous tags
	dangerousTags := []string{"iframe", "object", "embed", "form", "input", "button", "select", "textarea"}
	for _, tag := range dangerousTags {
		tagRegex := regexp.MustCompile(`(?i)</?` + tag + `[^>]*>`)
		htmlContent = tagRegex.ReplaceAllString(htmlContent, "")
	}

	// Remove javascript: and data: protocols from href and src attributes
	jsProtocolRegex := regexp.MustCompile(`(?i)(href|src)\s*=\s*["']?\s*(javascript|data):([^"'\s>]*)`)
	htmlContent = jsProtocolRegex.ReplaceAllString(htmlContent, `$1="#"`)

	// Remove on* event handlers
	eventRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	htmlContent = eventRegex.ReplaceAllString(htmlContent, "")

	return htmlContent
}

// preprocessMarkdown ensures proper spacing between headings and lists
func preprocessMarkdown(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var result []string

	for i, line := range lines {
		result = append(result, line)

		// Add empty line after headings if next line is a list item
		if i < len(lines)-1 {
			currentTrimmed := strings.TrimSpace(line)
			nextTrimmed := strings.TrimSpace(lines[i+1])

			// Check if current line is a heading
			isHeading := strings.HasPrefix(currentTrimmed, "#")

			// Check if next line is a list item
			isListItem := strings.HasPrefix(nextTrimmed, "- ") ||
				strings.HasPrefix(nextTrimmed, "* ") ||
				strings.HasPrefix(nextTrimmed, "+ ") ||
				regexp.MustCompile(`^\d+\. `).MatchString(nextTrimmed)

			// Add empty line between heading and list
			if isHeading && isListItem {
				result = append(result, "")
			}
		}
	}

	return strings.Join(result, "\n")
}
