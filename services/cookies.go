package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
)

type Cookies struct {
	*Base
}

func NewCookies(registry *Registry) *Cookies {
	return &Cookies{
		Base: NewBase(registry),
	}
}

func (c *Cookies) List() ([]*models.Cookie, *errors.MasterError) {
	rows, err := c.DB.Query(SQLListCookies)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanCookie)
}

func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, *errors.MasterError) {
	verr := c.validateApiKey(apiKey)
	if verr != nil {
		return nil, verr.MasterError()
	}

	rows, err := c.DB.Query(SQLListCookiesByApiKey, apiKey)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanCookie)
}

func (c *Cookies) Get(value string) (*models.Cookie, *errors.MasterError) {
	if value == "" {
		return nil, errors.NewMasterError(
			fmt.Errorf("cookie value missing"),
			http.StatusBadRequest,
		)
	}

	cookie, err := ScanCookie(c.DB.QueryRow(SQLGetCookieByValue, value))
	if err != nil {
		return cookie, errors.NewMasterError(err, 0)
	}

	return cookie, nil
}

func (c *Cookies) Add(userAgent, value, apiKey string) *errors.MasterError {
	verr := c.validateApiKey(apiKey)
	if verr != nil {
		return verr.MasterError()
	}

	// Check if cookie already exists
	count, merr := c.QueryCount(SQLCountCookies, value)
	if merr != nil {
		return merr
	}
	if count > 0 {
		return errors.NewValidationError("cookie for %s already exists",
			utils.MaskString(value)).MasterError()
	}

	lastLogin := time.Now().UnixMilli()
	_, err := c.DB.Exec(SQLAddCookie, userAgent, value, apiKey, lastLogin)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) Update(value string, cookie *models.Cookie) *errors.MasterError {
	verr := cookie.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	if value == "" {
		return errors.NewValidationError(
			"update cookie with value \"%s\" %s",
			utils.MaskString(value), cookie,
		).MasterError()
	}

	// First, get the current cookie to check its last_login timestamp
	currentCookie, dberr := c.Get(value)
	if dberr != nil {
		return dberr
	}

	// NOTE: Should i mutex lock here?
	_, err := c.DB.Exec(
		SQLUpdateCookie,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, // SET
		value,                   // WHERE ...
		currentCookie.LastLogin, // ... AND
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) Remove(value string) *errors.MasterError {
	if value == "" {
		return errors.NewMasterError(fmt.Errorf("value missing"), http.StatusBadRequest)
	}

	_, err := c.DB.Exec(SQLDeleteCookie, value)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) RemoveApiKey(apiKey string) *errors.MasterError {
	verr := c.validateApiKey(apiKey)
	if verr != nil {
		return verr.MasterError()
	}

	_, err := c.DB.Exec(SQLDeleteCookiesByApiKey, apiKey)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) RemoveExpired(beforeTimestamp int64) *errors.MasterError {
	_, err := c.DB.Exec(SQLDeleteExpiredCookies, beforeTimestamp)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) validateApiKey(key string) *errors.ValidationError {
	if !utils.ValidateAPIKey(key) {
		return errors.NewValidationError("invalid api_key")
	}
	return nil
}
