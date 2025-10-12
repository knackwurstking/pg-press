package tools

import (
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
)

func marshalFormat(format models.Format) ([]byte, error) {
	formatBytes, err := json.Marshal(format)
	if err != nil {
		return nil, fmt.Errorf("marshal error: tools: %v", err)
	}

	return formatBytes, nil
}
