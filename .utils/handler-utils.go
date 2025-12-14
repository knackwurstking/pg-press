package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services/shared"
	"github.com/labstack/echo/v4"
)

// SanitizeFilename cleans up a filename by replacing characters and removing duplicates
func SanitizeFilename(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}

	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "-", "_")
	filename = strings.ReplaceAll(filename, "(", "_")
	filename = strings.ReplaceAll(filename, ")", "_")
	filename = strings.ReplaceAll(filename, "[", "_")
	filename = strings.ReplaceAll(filename, "]", "_")

	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	filename = strings.Trim(filename, "_")

	if filename == "" {
		filename = "attachment"
	}

	return filename
}
