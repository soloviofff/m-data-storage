package migrations

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// calculateChecksum calculates SQL checksum
func calculateChecksum(sql string) string {
	hash := sha256.Sum256([]byte(sql))
	return fmt.Sprintf("%x", hash)
}

// LoadMigrationsFromFS loads migrations from filesystem
func LoadMigrationsFromFS(fsys fs.FS, dir string) ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		migration, err := parseMigrationFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to parse migration file %s: %w", path, err)
		}

		if migration != nil {
			migrations = append(migrations, *migration)
		}

		return nil
	})

	return migrations, err
}

// parseMigrationFile parses migration file
func parseMigrationFile(fsys fs.FS, path string) (*Migration, error) {
	// Extract version and name from filename
	// Expected format: 001_create_tables.up.sql or 001_create_tables.down.sql
	filename := filepath.Base(path)

	// Check if this is an up migration
	if !strings.Contains(filename, ".up.sql") {
		return nil, nil // Skip down files, they will be processed separately
	}

	version, name, err := parseFilename(filename)
	if err != nil {
		return nil, err
	}

	// Read up SQL
	upSQL, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read up SQL: %w", err)
	}

	// Look for corresponding down file
	downPath := strings.Replace(path, ".up.sql", ".down.sql", 1)
	var downSQL []byte
	if downFile, err := fs.ReadFile(fsys, downPath); err == nil {
		downSQL = downFile
	}

	migration := Migration{
		Version:     version,
		Name:        name,
		UpSQL:       string(upSQL),
		DownSQL:     string(downSQL),
		Description: fmt.Sprintf("Migration %d: %s", version, name),
	}

	return &migration, nil
}

// parseFilename extracts version and name from filename
func parseFilename(filename string) (int64, string, error) {
	// Regular expression for parsing filename
	// Format: 001_create_tables.up.sql
	re := regexp.MustCompile(`^(\d+)_(.+)\.up\.sql$`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid version number in filename: %s", filename)
	}

	name := strings.ReplaceAll(matches[2], "_", " ")

	return version, name, nil
}

// ValidateMigrations validates migration correctness
func ValidateMigrations(migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	// Check version uniqueness
	versions := make(map[int64]bool)
	names := make(map[string]bool)

	for _, migration := range migrations {
		// Check version uniqueness
		if versions[migration.Version] {
			return fmt.Errorf("duplicate migration version: %d", migration.Version)
		}
		versions[migration.Version] = true

		// Check name uniqueness
		if names[migration.Name] {
			return fmt.Errorf("duplicate migration name: %s", migration.Name)
		}
		names[migration.Name] = true

		// Check that version is positive
		if migration.Version <= 0 {
			return fmt.Errorf("migration version must be positive: %d", migration.Version)
		}

		// Check that up SQL exists
		if strings.TrimSpace(migration.UpSQL) == "" {
			return fmt.Errorf("migration %d (%s) has empty up SQL", migration.Version, migration.Name)
		}

		// Check name
		if strings.TrimSpace(migration.Name) == "" {
			return fmt.Errorf("migration %d has empty name", migration.Version)
		}
	}

	return nil
}

// GenerateMigrationFilename generates filename for new migration
func GenerateMigrationFilename(version int64, name string) (string, string) {
	// Replace spaces with underscores and convert to lowercase
	safeName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Remove invalid characters
	re := regexp.MustCompile(`[^a-z0-9_]`)
	safeName = re.ReplaceAllString(safeName, "")

	upFilename := fmt.Sprintf("%03d_%s.up.sql", version, safeName)
	downFilename := fmt.Sprintf("%03d_%s.down.sql", version, safeName)

	return upFilename, downFilename
}

// CreateMigrationTemplate creates migration template
func CreateMigrationTemplate(name, description string) (upSQL, downSQL string) {
	timestamp := time.Now().Format("2006-01-02")

	upSQL = fmt.Sprintf(`-- Migration: %s
-- Description: %s
-- Created: %s

-- Add your up migration SQL here

`, name, description, timestamp)

	downSQL = fmt.Sprintf(`-- Rollback for: %s
-- Description: %s
-- Created: %s

-- Add your down migration SQL here

`, name, description, timestamp)

	return upSQL, downSQL
}

// MigrationInfo contains migration information for display
type MigrationInfo struct {
	Version     int64  `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Applied     bool   `json:"applied"`
	AppliedAt   string `json:"applied_at,omitempty"`
	CanRollback bool   `json:"can_rollback"`
}

// GetMigrationInfo converts migration status to display information
func GetMigrationInfo(statuses []MigrationStatus) []MigrationInfo {
	info := make([]MigrationInfo, len(statuses))

	for i, status := range statuses {
		appliedAt := ""
		if status.Applied {
			appliedAt = status.AppliedAt.Format("2006-01-02 15:04:05")
		}

		info[i] = MigrationInfo{
			Version:     status.Migration.Version,
			Name:        status.Migration.Name,
			Description: status.Migration.Description,
			Applied:     status.Applied,
			AppliedAt:   appliedAt,
			CanRollback: status.Migration.DownSQL != "",
		}
	}

	return info
}
