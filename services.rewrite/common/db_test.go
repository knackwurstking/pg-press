package common

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	currentPath, err := os.Getwd()
	slog.Info("Current working directory: %s (err: %v)", currentPath, err)

	filesToRemove := []string{
		"userdb.sqlite",
		"userdb.sqlite-shm",
		"userdb.sqlite-wal",
	}

	for _, file := range filesToRemove {
		err = os.Remove(filepath.Join(currentPath, file))
		if err != nil {
			slog.Debug("Could not remove existing %s file: %v", file, err)
		}
	}
}

func TestNewDB(t *testing.T) {
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

func TestDBSetupAndClose(t *testing.T) {
	// Test that DB setup and close work correctly
	db := NewDB()
	assert.NotNil(t, db)

	// This should not error
	err := db.Setup()
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}

	db.Close()
}
