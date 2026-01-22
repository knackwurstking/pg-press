package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GetAttachmentFileName(fileName string) string {
	return fmt.Sprintf("%s%d%s",
		time.Now().Format("20060102150405"),
		uuid.New().ID(),
		strings.ToLower(filepath.Ext(fileName)),
	)
}
