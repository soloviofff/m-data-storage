package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a single migration
type Migration struct {
	Version     int64
	Name        string
	UpSQL       string
	DownSQL     string
	Description string
}

// MigrationRecord represents a record of executed migration
type MigrationRecord struct {
	Version   int64     `json:"version"`
	Name      string    `json:"name"`
	AppliedAt time.Time `json:"applied_at"`
	Checksum  string    `json:"checksum"`
}

// Migrator manages database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
	tableName  string
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB, tableName string) *Migrator {
	if tableName == "" {
		tableName = "schema_migrations"
	}

	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
		tableName:  tableName,
	}
}

// AddMigration adds migration to the list
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})
}

// AddMigrations adds multiple migrations
func (m *Migrator) AddMigrations(migrations []Migration) {
	for _, migration := range migrations {
		m.AddMigration(migration)
	}
}

// Up executes all unapplied migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	for _, migration := range m.migrations {
		if _, applied := appliedVersions[migration.Version]; !applied {
			if err := m.applyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d (%s): %w",
					migration.Version, migration.Name, err)
			}
		}
	}

	return nil
}

// Down rolls back migrations to specified version
func (m *Migrator) Down(ctx context.Context, targetVersion int64) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	// Get migrations in reverse order
	reversedMigrations := make([]Migration, len(m.migrations))
	copy(reversedMigrations, m.migrations)
	sort.Slice(reversedMigrations, func(i, j int) bool {
		return reversedMigrations[i].Version > reversedMigrations[j].Version
	})

	for _, migration := range reversedMigrations {
		if migration.Version <= targetVersion {
			break
		}

		if _, applied := appliedVersions[migration.Version]; applied {
			if err := m.rollbackMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to rollback migration %d (%s): %w",
					migration.Version, migration.Name, err)
			}
		}
	}

	return nil
}

// Status returns migration status
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied versions: %w", err)
	}

	status := make([]MigrationStatus, len(m.migrations))
	for i, migration := range m.migrations {
		record, applied := appliedVersions[migration.Version]
		status[i] = MigrationStatus{
			Migration: migration,
			Applied:   applied,
			AppliedAt: record.AppliedAt,
		}
	}

	return status, nil
}

// MigrationStatus represents migration status
type MigrationStatus struct {
	Migration Migration
	Applied   bool
	AppliedAt time.Time
}

// ensureMigrationsTable creates migrations table if it doesn't exist
func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			checksum TEXT NOT NULL
		)
	`, m.tableName)

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getAppliedVersions returns map of applied versions
func (m *Migrator) getAppliedVersions(ctx context.Context) (map[int64]MigrationRecord, error) {
	query := fmt.Sprintf(`
		SELECT version, name, applied_at, checksum 
		FROM %s 
		ORDER BY version
	`, m.tableName)

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int64]MigrationRecord)
	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Version, &record.Name, &record.AppliedAt, &record.Checksum); err != nil {
			return nil, err
		}
		versions[record.Version] = record
	}

	return versions, rows.Err()
}

// applyMigration applies migration
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration information
	checksum := calculateChecksum(migration.UpSQL)
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (version, name, checksum) 
		VALUES (?, ?, ?)
	`, m.tableName)

	if _, err := tx.ExecContext(ctx, insertQuery, migration.Version, migration.Name, checksum); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// rollbackMigration rolls back migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration Migration) error {
	if migration.DownSQL == "" {
		return fmt.Errorf("migration %d (%s) has no down SQL", migration.Version, migration.Name)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.ExecContext(ctx, migration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Delete migration record
	deleteQuery := fmt.Sprintf(`DELETE FROM %s WHERE version = ?`, m.tableName)
	if _, err := tx.ExecContext(ctx, deleteQuery, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return tx.Commit()
}
