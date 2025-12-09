package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDB(t *testing.T) {
	// Test that NewDB creates a valid DB instance
	db := NewDB()
	assert.NotNil(t, db)
	assert.NotNil(t, db.User)
	assert.NotNil(t, db.User.User)
	assert.NotNil(t, db.User.Cookie)
	assert.NotNil(t, db.User.Session)

	// Test setup functionality
	db.Setup()
	db.Close()
}

func TestDBSetupAndClose(t *testing.T) {
	// Test that DB setup and close work correctly
	db := NewDB()
	assert.NotNil(t, db)

	// This should not error
	db.Setup()
	db.Close()
}
