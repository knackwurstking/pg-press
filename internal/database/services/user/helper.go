package user

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Helper provides additional user-related database operations
// that are not part of the generic DataOperations interface.
type Helper struct {
	db *sql.DB
}

// NewHelper creates a new Helper instance.
func NewHelper(db *sql.DB) *Helper {
	return &Helper{
		db: db,
	}
}

// GetUserFromApiKey retrieves a user by their API key.
func (h *Helper) GetUserFromApiKey(apiKey string) (*models.User, error) {
	logger.DBUsers().Debug("Getting user by API key")

	if apiKey == "" {
		return nil, dberror.NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	query := `SELECT * FROM users WHERE api_key = ?`
	row := h.db.QueryRow(query, apiKey)

	user := &models.User{}
	err := row.Scan(&user.TelegramID, &user.UserName, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "users",
			"failed to get user by API key", err)
	}

	return user, nil
}
