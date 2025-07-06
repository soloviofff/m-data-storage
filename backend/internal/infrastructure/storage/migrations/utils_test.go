package migrations

import (
	"testing"
	"time"
)

func TestCalculateChecksum(t *testing.T) {
	sql1 := "CREATE TABLE test (id INTEGER);"
	sql2 := "CREATE TABLE test (id INTEGER);"
	sql3 := "CREATE TABLE test2 (id INTEGER);"

	checksum1 := calculateChecksum(sql1)
	checksum2 := calculateChecksum(sql2)
	checksum3 := calculateChecksum(sql3)

	// Same SQL should produce same checksum
	if checksum1 != checksum2 {
		t.Error("Same SQL should produce same checksum")
	}

	// Different SQL should produce different checksum
	if checksum1 == checksum3 {
		t.Error("Different SQL should produce different checksum")
	}

	// Checksum should be hex string
	if len(checksum1) != 64 {
		t.Errorf("Expected checksum length 64, got %d", len(checksum1))
	}
}

func TestParseFilename(t *testing.T) {
	tests := []struct {
		filename        string
		expectedVersion int64
		expectedName    string
		expectError     bool
	}{
		{"001_create_tables.up.sql", 1, "create tables", false},
		{"123_add_indexes.up.sql", 123, "add indexes", false},
		{"001_create_user_table.up.sql", 1, "create user table", false},
		{"invalid.sql", 0, "", true},
		{"001_test.down.sql", 0, "", true}, // down files should be ignored
		{"abc_test.up.sql", 0, "", true},   // invalid version
	}

	for _, test := range tests {
		version, name, err := parseFilename(test.filename)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error for filename %s, but got none", test.filename)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for filename %s: %v", test.filename, err)
			continue
		}

		if version != test.expectedVersion {
			t.Errorf("For filename %s, expected version %d, got %d", test.filename, test.expectedVersion, version)
		}

		if name != test.expectedName {
			t.Errorf("For filename %s, expected name '%s', got '%s'", test.filename, test.expectedName, name)
		}
	}
}

func TestValidateMigrations(t *testing.T) {
	// Valid migrations
	validMigrations := []Migration{
		{Version: 1, Name: "first", UpSQL: "CREATE TABLE test1 (id INTEGER);"},
		{Version: 2, Name: "second", UpSQL: "CREATE TABLE test2 (id INTEGER);"},
	}

	err := ValidateMigrations(validMigrations)
	if err != nil {
		t.Errorf("Valid migrations should not produce error: %v", err)
	}

	// Empty migrations should be valid
	err = ValidateMigrations([]Migration{})
	if err != nil {
		t.Errorf("Empty migrations should be valid: %v", err)
	}

	// Duplicate version
	duplicateVersionMigrations := []Migration{
		{Version: 1, Name: "first", UpSQL: "CREATE TABLE test1 (id INTEGER);"},
		{Version: 1, Name: "second", UpSQL: "CREATE TABLE test2 (id INTEGER);"},
	}

	err = ValidateMigrations(duplicateVersionMigrations)
	if err == nil {
		t.Error("Duplicate version should produce error")
	}

	// Duplicate name
	duplicateNameMigrations := []Migration{
		{Version: 1, Name: "test", UpSQL: "CREATE TABLE test1 (id INTEGER);"},
		{Version: 2, Name: "test", UpSQL: "CREATE TABLE test2 (id INTEGER);"},
	}

	err = ValidateMigrations(duplicateNameMigrations)
	if err == nil {
		t.Error("Duplicate name should produce error")
	}

	// Zero version
	zeroVersionMigrations := []Migration{
		{Version: 0, Name: "test", UpSQL: "CREATE TABLE test (id INTEGER);"},
	}

	err = ValidateMigrations(zeroVersionMigrations)
	if err == nil {
		t.Error("Zero version should produce error")
	}

	// Empty up SQL
	emptyUpSQLMigrations := []Migration{
		{Version: 1, Name: "test", UpSQL: ""},
	}

	err = ValidateMigrations(emptyUpSQLMigrations)
	if err == nil {
		t.Error("Empty up SQL should produce error")
	}

	// Empty name
	emptyNameMigrations := []Migration{
		{Version: 1, Name: "", UpSQL: "CREATE TABLE test (id INTEGER);"},
	}

	err = ValidateMigrations(emptyNameMigrations)
	if err == nil {
		t.Error("Empty name should produce error")
	}
}

func TestGenerateMigrationFilename(t *testing.T) {
	tests := []struct {
		version      int64
		name         string
		expectedUp   string
		expectedDown string
	}{
		{1, "create tables", "001_create_tables.up.sql", "001_create_tables.down.sql"},
		{123, "add indexes", "123_add_indexes.up.sql", "123_add_indexes.down.sql"},
		{1, "Create User Table", "001_create_user_table.up.sql", "001_create_user_table.down.sql"},
		{1, "test-migration", "001_testmigration.up.sql", "001_testmigration.down.sql"},
	}

	for _, test := range tests {
		upFilename, downFilename := GenerateMigrationFilename(test.version, test.name)

		if upFilename != test.expectedUp {
			t.Errorf("For version %d and name '%s', expected up filename '%s', got '%s'",
				test.version, test.name, test.expectedUp, upFilename)
		}

		if downFilename != test.expectedDown {
			t.Errorf("For version %d and name '%s', expected down filename '%s', got '%s'",
				test.version, test.name, test.expectedDown, downFilename)
		}
	}
}

func TestCreateMigrationTemplate(t *testing.T) {
	name := "test migration"
	description := "Test migration description"

	upSQL, downSQL := CreateMigrationTemplate(name, description)

	// Check that templates contain expected content
	if !contains(upSQL, name) {
		t.Error("Up SQL template should contain migration name")
	}

	if !contains(upSQL, description) {
		t.Error("Up SQL template should contain description")
	}

	if !contains(downSQL, name) {
		t.Error("Down SQL template should contain migration name")
	}

	if !contains(downSQL, description) {
		t.Error("Down SQL template should contain description")
	}

	// Check that templates contain current date
	currentDate := time.Now().Format("2006-01-02")
	if !contains(upSQL, currentDate) {
		t.Error("Up SQL template should contain current date")
	}

	if !contains(downSQL, currentDate) {
		t.Error("Down SQL template should contain current date")
	}
}

func TestGetMigrationInfo(t *testing.T) {
	appliedTime := time.Now()
	statuses := []MigrationStatus{
		{
			Migration: Migration{
				Version:     1,
				Name:        "first migration",
				Description: "First test migration",
				DownSQL:     "DROP TABLE test;",
			},
			Applied:   true,
			AppliedAt: appliedTime,
		},
		{
			Migration: Migration{
				Version:     2,
				Name:        "second migration",
				Description: "Second test migration",
				DownSQL:     "",
			},
			Applied:   false,
			AppliedAt: time.Time{},
		},
	}

	info := GetMigrationInfo(statuses)

	if len(info) != 2 {
		t.Errorf("Expected 2 migration info items, got %d", len(info))
	}

	// Check first migration info
	if info[0].Version != 1 {
		t.Errorf("Expected version 1, got %d", info[0].Version)
	}
	if !info[0].Applied {
		t.Error("Expected first migration to be applied")
	}
	if !info[0].CanRollback {
		t.Error("Expected first migration to be rollbackable")
	}
	if info[0].AppliedAt == "" {
		t.Error("Expected applied time to be set")
	}

	// Check second migration info
	if info[1].Version != 2 {
		t.Errorf("Expected version 2, got %d", info[1].Version)
	}
	if info[1].Applied {
		t.Error("Expected second migration to not be applied")
	}
	if info[1].CanRollback {
		t.Error("Expected second migration to not be rollbackable")
	}
	if info[1].AppliedAt != "" {
		t.Error("Expected applied time to be empty")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 1; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}
