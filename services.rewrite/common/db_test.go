package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDB(t *testing.T) {
	currentPath, err := os.Getwd()
	t.Logf("Current working directory: %s (err: %v)", currentPath, err)

	err = os.Remove(filepath.Join(currentPath, "userdb.sqlite"))
	if err != nil {
		t.Logf("Could not remove existing userdb.sqlite file: %v", err)
	}

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
