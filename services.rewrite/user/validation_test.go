package user

import (
	"testing"

	"github.com/knackwurstking/pg-press/services.rewrite/common"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
	"github.com/stretchr/testify/assert"
)

func TestUserServiceValidation(t *testing.T) {
	// Initialize DB
	db := common.NewDB()
	defer db.Close()

	// Test creating a user with invalid data
	user := &shared.User{
		Name:    "", // Empty name should fail validation
		ApiKey:  "test-api-key",
		LastFeed: "test-feed",
	}

	err := db.User.User.Create(user)
	assert.Error(t, err)

	// Test getting a user with invalid ID
	_, err = db.User.User.GetByID(0)
	assert.Error(t, err)

	// Test updating a user with invalid data
	user.ID = 1
	user.Name = ""
	err = db.User.User.Update(user)
	assert.Error(t, err)
}

func TestCookieServiceValidation(t *testing.T) {
	// Initialize DB
	db := common.NewDB()
	defer db.Close()

	// Test creating a cookie with invalid data
	cookie := &shared.Cookie{
		UserAgent: "Test User Agent",
		Value:     "", // Empty value should fail validation
		UserID:    1,
		LastLogin: 1234567890,
	}

	err := db.User.Cookie.Create(cookie)
	assert.Error(t, err)

	// Test getting a cookie with empty value
	_, err = db.User.Cookie.GetByID("")
	assert.Error(t, err)

	// Test deleting a cookie with empty value
	err = db.User.Cookie.Delete("")
	assert.Error(t, err)
}

func TestSessionServiceValidation(t *testing.T) {
	// Initialize DB
	db := common.NewDB()
	defer db.Close()

	// Test creating a session with invalid data
	session := &shared.Session{
		ID:       0, // Invalid ID should fail validation
		UserID:   1,
		Expiry:   1234567890,
		Active:   true,
		Data:     "test session data",
		IP:       "127.0.0.1",
		UserAgent: "Test User Agent",
	}

	err := db.User.Session.Create(session)
	assert.Error(t, err)

	// Test getting a session with invalid ID
	_, err = db.User.Session.GetByID(0)
	assert.Error(t, err)

	// Test deleting a session with invalid ID
	err = db.User.Session.Delete(0)
	assert.Error(t, err)
}

