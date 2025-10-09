package attachments

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func scanAttachmentsFromRows(rows *sql.Rows) ([]*models.Attachment, error) {
	return scanner.ScanRows(rows, scanAttachment)
}

func scanAttachment(scanner interfaces.Scannable) (*models.Attachment, error) {
	attachment := &models.Attachment{}
	var id int64

	err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan attachment: %v", err)
	}

	// Set the ID using string conversion to maintain compatibility
	attachment.ID = fmt.Sprintf("%d", id)
	return attachment, nil
}

func scanAttachmentsIntoMap(rows *sql.Rows) (map[int64]*models.Attachment, error) {
	return scanner.ScanIntoMap(rows, scanAttachment, func(attachment *models.Attachment) int64 {
		// Convert string ID back to int64 for map key
		id := int64(0)
		fmt.Sscanf(attachment.ID, "%d", &id)
		return id
	})
}
