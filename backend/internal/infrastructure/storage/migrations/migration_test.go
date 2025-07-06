package migrations

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database
	dbFile := "test_migrations.db"
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(dbFile)
	}

	return db, cleanup
}

func TestMigrator_AddMigration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	migrator := NewMigrator(db, "test_migrations")

	migration := Migration{
		Version: 1,
		Name:    "test migration",
		UpSQL:   "CREATE TABLE test (id INTEGER PRIMARY KEY);",
		DownSQL: "DROP TABLE test;",
	}

	migrator.AddMigration(migration)

	if len(migrator.migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(migrator.migrations))
	}

	if migrator.migrations[0].Version != 1 {
		t.Errorf("Expected version 1, got %d", migrator.migrations[0].Version)
	}
}

func TestMigrator_Up(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	migrator := NewMigrator(db, "test_migrations")
	ctx := context.Background()

	// Add test migration
	migration := Migration{
		Version: 1,
		Name:    "create test table",
		UpSQL:   "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);",
		DownSQL: "DROP TABLE test_table;",
	}
	migrator.AddMigration(migration)

	// Run migration
	err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check that table was created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table was not created: %v", err)
	}

	if tableName != "test_table" {
		t.Errorf("Expected table name 'test_table', got '%s'", tableName)
	}

	// Check migration record
	var version int64
	var name string
	err = db.QueryRow("SELECT version, name FROM test_migrations WHERE version = 1").Scan(&version, &name)
	if err != nil {
		t.Fatalf("Migration record not found: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
	if name != "create test table" {
		t.Errorf("Expected name 'create test table', got '%s'", name)
	}
}

func TestMigrator_Down(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	migrator := NewMigrator(db, "test_migrations")
	ctx := context.Background()

	// Add test migrations
	migration1 := Migration{
		Version: 1,
		Name:    "create test table",
		UpSQL:   "CREATE TABLE test_table (id INTEGER PRIMARY KEY);",
		DownSQL: "DROP TABLE test_table;",
	}
	migration2 := Migration{
		Version: 2,
		Name:    "add column",
		UpSQL:   "ALTER TABLE test_table ADD COLUMN name TEXT;",
		DownSQL: "ALTER TABLE test_table DROP COLUMN name;",
	}

	migrator.AddMigration(migration1)
	migrator.AddMigration(migration2)

	// Run migrations up
	err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Migration up failed: %v", err)
	}

	// Run migration down to version 1
	err = migrator.Down(ctx, 1)
	if err != nil {
		t.Fatalf("Migration down failed: %v", err)
	}

	// Check that only migration 1 is applied
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count migrations: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 migration record, got %d", count)
	}

	var version int64
	err = db.QueryRow("SELECT version FROM test_migrations").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to get migration version: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
}

func TestMigrator_Status(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	migrator := NewMigrator(db, "test_migrations")
	ctx := context.Background()

	// Add test migrations
	migration1 := Migration{
		Version: 1,
		Name:    "first migration",
		UpSQL:   "CREATE TABLE test1 (id INTEGER);",
		DownSQL: "DROP TABLE test1;",
	}
	migration2 := Migration{
		Version: 2,
		Name:    "second migration",
		UpSQL:   "CREATE TABLE test2 (id INTEGER);",
		DownSQL: "DROP TABLE test2;",
	}

	migrator.AddMigration(migration1)
	migrator.AddMigration(migration2)

	// Ensure migrations table exists
	err := migrator.ensureMigrationsTable(ctx)
	if err != nil {
		t.Fatalf("Failed to ensure migrations table: %v", err)
	}

	// Apply first migration only
	err = migrator.applyMigration(ctx, migration1)
	if err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Get status
	status, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status) != 2 {
		t.Errorf("Expected 2 migrations in status, got %d", len(status))
	}

	// Check first migration status
	if !status[0].Applied {
		t.Error("Expected first migration to be applied")
	}
	if status[0].Migration.Version != 1 {
		t.Errorf("Expected first migration version 1, got %d", status[0].Migration.Version)
	}

	// Check second migration status
	if status[1].Applied {
		t.Error("Expected second migration to not be applied")
	}
	if status[1].Migration.Version != 2 {
		t.Errorf("Expected second migration version 2, got %d", status[1].Migration.Version)
	}
}

func TestMigrator_EnsureMigrationsTable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	migrator := NewMigrator(db, "test_migrations")
	ctx := context.Background()

	// Ensure migrations table
	err := migrator.ensureMigrationsTable(ctx)
	if err != nil {
		t.Fatalf("Failed to ensure migrations table: %v", err)
	}

	// Check that table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_migrations'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Migrations table was not created: %v", err)
	}

	if tableName != "test_migrations" {
		t.Errorf("Expected table name 'test_migrations', got '%s'", tableName)
	}
}

func TestMigrator_SortMigrations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	migrator := NewMigrator(db, "test_migrations")

	// Add migrations in random order
	migration3 := Migration{Version: 3, Name: "third", UpSQL: "SELECT 3;"}
	migration1 := Migration{Version: 1, Name: "first", UpSQL: "SELECT 1;"}
	migration2 := Migration{Version: 2, Name: "second", UpSQL: "SELECT 2;"}

	migrator.AddMigration(migration3)
	migrator.AddMigration(migration1)
	migrator.AddMigration(migration2)

	// Check that migrations are sorted by version
	if migrator.migrations[0].Version != 1 {
		t.Errorf("Expected first migration version 1, got %d", migrator.migrations[0].Version)
	}
	if migrator.migrations[1].Version != 2 {
		t.Errorf("Expected second migration version 2, got %d", migrator.migrations[1].Version)
	}
	if migrator.migrations[2].Version != 3 {
		t.Errorf("Expected third migration version 3, got %d", migrator.migrations[2].Version)
	}
}
