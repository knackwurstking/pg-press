package user

import (
	"testing"

	"github.com/knackwurstking/pg-press/services.rewrite/common"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
	"github.com/stretchr/testify/assert"
)

func TestUserService(t *testing.T) {
	// Initialize DB
	db := common.NewDB()
	defer db.Close()

	// Test creating a user
	user := &shared.User{
		Name:     "Test User",
		ApiKey:   "test-api-key",
		LastFeed: 0,
	}

	err := db.User.User.Create(user)
	assert.NoError(t, err)
	assert.Greater(t, user.ID, shared.TelegramID(0))

	// Test getting a user by ID
	fetchedUser, err := db.User.User.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, fetchedUser.Name)
	assert.Equal(t, user.ApiKey, fetchedUser.ApiKey)
	assert.Equal(t, user.LastFeed, fetchedUser.LastFeed)

	// Test updating a user
	user.Name = "Updated Test User"
	err = db.User.User.Update(user)
	assert.NoError(t, err)

	// Verify update
	updatedUser, err := db.User.User.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, updatedUser.Name)

	// Test listing users
	users, err := db.User.User.List()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 1)

	// Test deleting a user
	err = db.User.User.Delete(user.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = db.User.User.GetByID(user.ID)
	assert.Error(t, err)
}

func TestCookieService(t *testing.T) {
	// Initialize DB
	db := common.NewDB()
	defer db.Close()

	// Test creating a cookie
	cookie := &shared.Cookie{
		UserAgent: "Test User Agent",
		Value:     "test-cookie-value",
		UserID:    1,
		LastLogin: 1234567890,
	}

	err := db.User.Cookie.Create(cookie)
	assert.NoError(t, err)

	// Test getting a cookie by ID
	fetchedCookie, err := db.User.Cookie.GetByID(cookie.Value)
	assert.NoError(t, err)
	assert.Equal(t, cookie.UserAgent, fetchedCookie.UserAgent)
	assert.Equal(t, cookie.Value, fetchedCookie.Value)
	assert.Equal(t, cookie.UserID, fetchedCookie.UserID)
	assert.Equal(t, cookie.LastLogin, fetchedCookie.LastLogin)

	// Test updating a cookie
	cookie.UserAgent = "Updated User Agent"
	err = db.User.Cookie.Update(cookie)
	assert.NoError(t, err)

	// Verify update
	updatedCookie, err := db.User.Cookie.GetByID(cookie.Value)
	assert.NoError(t, err)
	assert.Equal(t, cookie.UserAgent, updatedCookie.UserAgent)

	// Test listing cookies
	cookies, err := db.User.Cookie.List()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(cookies), 1)

	// Test deleting a cookie
	err = db.User.Cookie.Delete(cookie.Value)
	assert.NoError(t, err)

	// Verify deletion
	_, err = db.User.Cookie.GetByID(cookie.Value)
	assert.Error(t, err)
}

func TestSessionService(t *testing.T) {
	// Initialize DB
	db := common.NewDB()
	defer db.Close()

	// Test creating a session
	session := &shared.Session{
		ID:        1,
		UserID:    1,
		Expiry:    1234567890,
		Active:    true,
		Data:      "test session data",
		IP:        "127.0.0.1",
		UserAgent: "Test User Agent",
	}

	err := db.User.Session.Create(session)
	assert.NoError(t, err)

	// Test getting a session by ID
	fetchedSession, err := db.User.Session.GetByID(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.UserID, fetchedSession.UserID)
	assert.Equal(t, session.Expiry, fetchedSession.Expiry)
	assert.Equal(t, session.Active, fetchedSession.Active)
	assert.Equal(t, session.Data, fetchedSession.Data)
	assert.Equal(t, session.IP, fetchedSession.IP)
	assert.Equal(t, session.UserAgent, fetchedSession.UserAgent)

	// Test updating a session
	session.Data = "updated session data"
	err = db.User.Session.Update(session)
	assert.NoError(t, err)

	// Verify update
	updatedSession, err := db.User.Session.GetByID(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.Data, updatedSession.Data)

	// Test listing sessions
	sessions, err := db.User.Session.List()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(sessions), 1)

	// Test deleting a session
	err = db.User.Session.Delete(session.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = db.User.Session.GetByID(session.ID)
	assert.Error(t, err)
}

func TestDBInitialization(t *testing.T) {
	// Test that DB initializes properly
	db := common.NewDB()
	assert.NotNil(t, db)
	assert.NotNil(t, db.User)
	assert.NotNil(t, db.User.User)
	assert.NotNil(t, db.User.Cookie)
	assert.NotNil(t, db.User.Session)

	// Test that setup works
	db.Setup()
	db.Close()
}
