package user

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
)

// UsersHelper provides additional user-related database operations
// that are not part of the generic DataOperations interface.
type UsersHelper struct {
	db *sql.DB
}

// NewUsersHelper creates a new UsersHelper instance.
func NewUsersHelper(db *sql.DB) *UsersHelper {
	return &UsersHelper{
		db: db,
	}
}

// GetUserFromApiKey retrieves a user by their API key.
func (uh *UsersHelper) GetUserFromApiKey(apiKey string) (*models.User, error) {
	logger.DBUsers().Debug("Getting user by API key")

	if apiKey == "" {
		return nil, dberror.NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	query := `SELECT * FROM users WHERE api_key = ?`
	row := uh.db.QueryRow(query, apiKey)

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
