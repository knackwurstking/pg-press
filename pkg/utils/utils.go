package utils

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

type ValidationError struct {
	message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (v *ValidationError) Error() string {
	return "validation error: " + v.message
}

func IsNotValidationError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*ValidationError); ok {
		return true
	}

	return false
}

type NotFoundError struct {
	message string
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message: message}
}

func (nf *NotFoundError) Error() string {
	return "not found: " + nf.message
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*NotFoundError); ok {
		return true
	}

	return false
}

type AlreadyExistsError struct {
	message string
}

func NewAlreadyExistsError(message string) *AlreadyExistsError {
	return &AlreadyExistsError{message: message}
}

func (ae *AlreadyExistsError) Error() string {
	return "already exists: " + ae.message
}

func IsAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*AlreadyExistsError); ok {
		return true
	}

	return false
}

type InvalidCredentialsError struct {
	message string
}

func NewInvalidCredentialsError(message string) *InvalidCredentialsError {
	return &InvalidCredentialsError{message: message}
}

func (ic *InvalidCredentialsError) Error() string {
	return "invalid credentials: " + ic.message
}

func IsInvalidCredentialsError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*InvalidCredentialsError); ok {
		return true
	}

	return false
}

func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if IsNotFoundError(err) {
		return http.StatusNotFound
	}

	if IsAlreadyExistsError(err) {
		return http.StatusConflict
	}

	if IsInvalidCredentialsError(err) {
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}

// MaskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
func MaskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
