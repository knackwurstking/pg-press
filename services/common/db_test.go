package common

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/williepotgieter/keymaker"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

func TestNewDB(t *testing.T) {
	removeDBFiles()

	// Test that NewDB creates a valid DB instance
	db := NewDB()
	assert.NotNil(t, db)
	assert.NotNil(t, db.User)
	assert.NotNil(t, db.User.User)
	assert.NotNil(t, db.User.Cookie)
	assert.NotNil(t, db.User.Session)

	// Test setup functionality
	err := db.Setup()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	db.Close()
}

func TestUserCRUD(t *testing.T) {
	removeDBFiles()

	db := NewDB()
	assert.NotNil(t, db)

	err := db.Setup()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	defer db.Close()

	// Create a test user
	apiKey, _ := keymaker.NewApiKey("pgp", 32)
	userEntity := &shared.User{
		Name:   "Test User",
		ApiKey: apiKey.String(),
	}

	// Test create user
	merr := db.User.User.Create(userEntity)
	assertNoMasterError(t, merr, "Error creating user")
	assert.True(t, userEntity.ID > 0)

	// Test get user by ID
	fetchedUser, merr := db.User.User.GetByID(userEntity.ID)
	assertNoMasterError(t, merr, "Error fetching user by ID")
	assert.NotNil(t, fetchedUser)
	assert.Equal(t, userEntity.Name, fetchedUser.Name)
	assert.Equal(t, userEntity.ApiKey, fetchedUser.ApiKey)

	// Test update user
	fetchedUser.Name = "Updated Test User"
	merr = db.User.User.Update(fetchedUser)
	assertNoMasterError(t, merr, "Error updating user")

	// Verify update
	updatedUser, merr := db.User.User.GetByID(userEntity.ID)
	assertNoMasterError(t, merr, "Error fetching updated user")
	assert.Equal(t, "Updated Test User", updatedUser.Name)

	// Test list users
	users, merr := db.User.User.List()
	assertNoMasterError(t, merr, "Error listing users")
	assert.Len(t, users, 1)

	// Test delete user
	merr = db.User.User.Delete(userEntity.ID)
	assertNoMasterError(t, merr, "Error deleting user")

	// Verify deletion, expecting an error
	_, merr = db.User.User.GetByID(userEntity.ID)
	if merr == nil {
		t.Fatalf("Expected error when fetching deleted user, got none")
	}
}

func TestCookieCRUD(t *testing.T) {
	removeDBFiles()

	db := NewDB()
	assert.NotNil(t, db)

	err := db.Setup()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	defer db.Close()

	// Create a test cookie
	cookieEntity := &shared.Cookie{
		UserAgent: "test-agent",
		Value:     "test-cookie-value",
		UserID:    1,
		LastLogin: 1234567890,
	}

	// Test create cookie
	merr := db.User.Cookie.Create(cookieEntity)
	assertNoMasterError(t, merr, "Error creating cookie")

	// Test get cookie by value
	fetchedCookie, merr := db.User.Cookie.GetByID(cookieEntity.Value)
	assertNoMasterError(t, merr, "Error fetching cookie by value")
	assert.NotNil(t, fetchedCookie)
	assert.Equal(t, cookieEntity.UserAgent, fetchedCookie.UserAgent)
	assert.Equal(t, cookieEntity.UserID, fetchedCookie.UserID)

	// Test update cookie
	fetchedCookie.UserAgent = "updated-agent"
	merr = db.User.Cookie.Update(fetchedCookie)
	assertNoMasterError(t, merr, "Error updating cookie")

	// Verify update
	updatedCookie, merr := db.User.Cookie.GetByID(cookieEntity.Value)
	assertNoMasterError(t, merr, "Error fetching updated cookie")
	assert.Equal(t, "updated-agent", updatedCookie.UserAgent)

	// Test list cookies
	cookies, merr := db.User.Cookie.List()
	assertNoMasterError(t, merr, "Error listing cookies")
	assert.Len(t, cookies, 1)

	// Test delete cookie
	merr = db.User.Cookie.Delete(cookieEntity.Value)
	assertNoMasterError(t, merr, "Error deleting cookie")

	// Verify deletion, expecting an error
	_, merr = db.User.Cookie.GetByID(cookieEntity.Value)
	if merr == nil {
		t.Fatalf("Expected error when fetching deleted cookie, got none")
	}
}

func TestSessionCRUD(t *testing.T) {
	removeDBFiles()

	db := NewDB()
	assert.NotNil(t, db)

	err := db.Setup()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	defer db.Close()

	// Create a test session
	sessionEntity := &shared.Session{
		ID: 1,
	}

	// Test create session
	merr := db.User.Session.Create(sessionEntity)
	assertNoMasterError(t, merr, "Error creating session")

	// Test get session by ID
	fetchedSession, merr := db.User.Session.GetByID(sessionEntity.ID)
	assertNoMasterError(t, merr, "Error fetching session by ID")
	assert.NotNil(t, fetchedSession)
	assert.Equal(t, sessionEntity.ID, fetchedSession.ID)

	// Test update session
	merr = db.User.Session.Update(sessionEntity)
	assertNoMasterError(t, merr, "Error updating session")

	// Verify update
	updatedSession, merr := db.User.Session.GetByID(sessionEntity.ID)
	assertNoMasterError(t, merr, "Error fetching updated session")
	assert.Equal(t, sessionEntity.ID, updatedSession.ID)

	// Test list sessions
	sessions, merr := db.User.Session.List()
	assertNoMasterError(t, merr, "Error listing sessions")
	assert.Len(t, sessions, 1)

	// Test delete session
	merr = db.User.Session.Delete(sessionEntity.ID)
	assertNoMasterError(t, merr, "Error deleting session")

	// Verify deletion, expecting an error
	_, merr = db.User.Session.GetByID(sessionEntity.ID)
	if merr == nil {
		t.Fatalf("Expected error when fetching deleted session, got none")
	}
}

func removeDBFiles() {
	currentPath, _ := os.Getwd()
	filesToRemove := []string{
		"userdb.sqlite",
		"userdb.sqlite-shm",
		"userdb.sqlite-wal",
	}

	for _, file := range filesToRemove {
		_ = os.Remove(filepath.Join(currentPath, file))
	}
}

func assertNoMasterError(t *testing.T, merr *errors.MasterError, format string, args ...any) {
	if merr != nil {
		t.Fatalf("%s: %v", fmt.Sprintf(format, args...), merr)
	}
}
