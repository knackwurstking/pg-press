package cookies

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
}

func NewService(db *sql.DB) *Service {
	base := base.NewBaseService(db, "Cookies")

	query := `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`

	if err := base.CreateTable(query, "cookies"); err != nil {
		panic(err)
	}

	return &Service{
		BaseService: base,
	}
}

func (c *Service) List() ([]*models.Cookie, error) {
	c.Log.Debug("Listing cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, c.HandleSelectError(err, "cookies")
	}
	defer rows.Close()

	cookies, err := scanCookiesFromRows(rows)
	if err != nil {
		return nil, err
	}

	c.Log.Debug("Listed cookies: count: %d", len(cookies))
	return cookies, nil
}

func (c *Service) ListApiKey(apiKey string) ([]*models.Cookie, error) {
	if err := validation.ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	c.Log.Debug("Listing cookies by API key")

	query := `SELECT * FROM cookies WHERE api_key = ? ORDER BY last_login DESC`
	rows, err := c.DB.Query(query, apiKey)
	if err != nil {
		return nil, c.HandleSelectError(err, "cookies")
	}
	defer rows.Close()

	cookies, err := scanCookiesFromRows(rows)
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

func (c *Service) Get(value string) (*models.Cookie, error) {
	if err := validation.ValidateNotEmpty(value, "value"); err != nil {
		return nil, err
	}

	c.Log.Debug("Getting cookie by value")

	query := `SELECT * FROM cookies WHERE value = ?`
	row := c.DB.QueryRow(query, value)

	cookie, err := scanner.ScanSingleRow(row, scanCookie, "cookies")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(value)
		}
		return nil, err
	}

	return cookie, nil
}

func (c *Service) Add(cookie *models.Cookie) error {
	if err := validateCookie(cookie); err != nil {
		return err
	}

	c.Log.Debug("Adding cookie")

	// Check if cookie already exists
	exists, err := c.CheckExistence(`SELECT COUNT(*) FROM cookies WHERE value = ?`, cookie.Value)
	if err != nil {
		return c.HandleSelectError(err, "cookies")
	}

	if exists {
		return utils.NewAlreadyExistsError("cookie already exists")
	}

	query := `INSERT INTO cookies (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`
	_, err = c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return c.HandleInsertError(err, "cookies")
	}

	return nil
}

func (c *Service) Update(value string, cookie *models.Cookie) error {
	if err := validation.ValidateNotEmpty(value, "current_value"); err != nil {
		return err
	}

	if err := validateCookie(cookie); err != nil {
		return err
	}

	c.Log.Debug("Updating cookie")

	query := `UPDATE cookies SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`
	result, err := c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return c.HandleUpdateError(err, "cookies")
	}

	if err := c.CheckRowsAffected(result, "cookie", value); err != nil {
		return err
	}

	return nil
}

func (c *Service) Remove(value string) error {
	if err := validation.ValidateNotEmpty(value, "value"); err != nil {
		return err
	}

	c.Log.Debug("Removing cookie")

	query := `DELETE FROM cookies WHERE value = ?`
	result, err := c.DB.Exec(query, value)
	if err != nil {
		return c.HandleDeleteError(err, "cookies")
	}

	if err := c.CheckRowsAffected(result, "cookie", value); err != nil {
		return err
	}

	return nil
}

func (c *Service) RemoveApiKey(apiKey string) error {
	if err := validation.ValidateAPIKey(apiKey); err != nil {
		return err
	}

	c.Log.Debug("Removing cookies by API key")

	query := `DELETE FROM cookies WHERE api_key = ?`
	result, err := c.DB.Exec(query, apiKey)
	if err != nil {
		return c.HandleDeleteError(err, "cookies")
	}

	rowsAffected, err := c.GetRowsAffected(result, "remove cookies by API key")
	if err != nil {
		return err
	}

	c.Log.Debug("Successfully removed cookies for API key: count: %d", rowsAffected)
	return nil
}

func (c *Service) RemoveExpired(beforeTimestamp int64) (int64, error) {
	if err := validation.ValidateTimestamp(beforeTimestamp, "timestamp"); err != nil {
		return 0, err
	}

	c.Log.Debug("Removing expired cookies, before_timestamp: %d", beforeTimestamp)

	query := `DELETE FROM cookies WHERE last_login < ?`
	result, err := c.DB.Exec(query, beforeTimestamp)
	if err != nil {
		return 0, c.HandleDeleteError(err, "cookies")
	}

	rowsAffected, err := c.GetRowsAffected(result, "remove expired cookies")
	if err != nil {
		return 0, err
	}

	c.Log.Debug("Removed expired cookies: count: %d", rowsAffected)
	return rowsAffected, nil
}
