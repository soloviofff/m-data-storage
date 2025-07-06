package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewManager(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	if manager.metadata == nil {
		t.Error("Metadata storage is nil")
	}

	if manager.timeSeries == nil {
		t.Error("Time series storage is nil")
	}

	if manager.migrations != nil {
		t.Error("Migration manager should be nil before initialization")
	}

	if manager.logger == nil {
		t.Error("Logger is nil")
	}
}

func TestNewManagerWithNilLogger(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.logger == nil {
		t.Error("Logger should be created when nil is passed")
	}
}

func TestManagerMigrationMethods(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test methods before initialization - should fail
	status, err := manager.GetMigrationStatus(ctx)
	if err == nil {
		t.Error("GetMigrationStatus should fail before initialization")
	}
	if status != nil {
		t.Error("Migration status should be nil before initialization")
	}

	err = manager.RunMigrations(ctx)
	if err == nil {
		t.Error("RunMigrations should fail before initialization")
	}

	err = manager.RollbackMigrations(ctx, 0)
	if err == nil {
		t.Error("RollbackMigrations should fail before initialization")
	}

	t.Log("All migration methods correctly fail before initialization")
}
