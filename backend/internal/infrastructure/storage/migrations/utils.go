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

// calculateChecksum вычисляет контрольную сумму SQL
func calculateChecksum(sql string) string {
	hash := sha256.Sum256([]byte(sql))
	return fmt.Sprintf("%x", hash)
}

// LoadMigrationsFromFS загружает миграции из файловой системы
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

// parseMigrationFile парсит файл миграции
func parseMigrationFile(fsys fs.FS, path string) (*Migration, error) {
	// Извлекаем версию и имя из имени файла
	// Ожидаемый формат: 001_create_tables.up.sql или 001_create_tables.down.sql
	filename := filepath.Base(path)

	// Проверяем, является ли это up миграцией
	if !strings.Contains(filename, ".up.sql") {
		return nil, nil // Пропускаем down файлы, они будут обработаны отдельно
	}

	version, name, err := parseFilename(filename)
	if err != nil {
		return nil, err
	}

	// Читаем up SQL
	upSQL, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read up SQL: %w", err)
	}

	// Ищем соответствующий down файл
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

// parseFilename извлекает версию и имя из имени файла
func parseFilename(filename string) (int64, string, error) {
	// Регулярное выражение для парсинга имени файла
	// Формат: 001_create_tables.up.sql
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

// ValidateMigrations проверяет корректность миграций
func ValidateMigrations(migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	// Проверяем уникальность версий
	versions := make(map[int64]bool)
	names := make(map[string]bool)

	for _, migration := range migrations {
		// Проверяем уникальность версии
		if versions[migration.Version] {
			return fmt.Errorf("duplicate migration version: %d", migration.Version)
		}
		versions[migration.Version] = true

		// Проверяем уникальность имени
		if names[migration.Name] {
			return fmt.Errorf("duplicate migration name: %s", migration.Name)
		}
		names[migration.Name] = true

		// Проверяем, что версия положительная
		if migration.Version <= 0 {
			return fmt.Errorf("migration version must be positive: %d", migration.Version)
		}

		// Проверяем, что есть up SQL
		if strings.TrimSpace(migration.UpSQL) == "" {
			return fmt.Errorf("migration %d (%s) has empty up SQL", migration.Version, migration.Name)
		}

		// Проверяем имя
		if strings.TrimSpace(migration.Name) == "" {
			return fmt.Errorf("migration %d has empty name", migration.Version)
		}
	}

	return nil
}

// GenerateMigrationFilename генерирует имя файла для новой миграции
func GenerateMigrationFilename(version int64, name string) (string, string) {
	// Заменяем пробелы на подчеркивания и приводим к нижнему регистру
	safeName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// Удаляем недопустимые символы
	re := regexp.MustCompile(`[^a-z0-9_]`)
	safeName = re.ReplaceAllString(safeName, "")

	upFilename := fmt.Sprintf("%03d_%s.up.sql", version, safeName)
	downFilename := fmt.Sprintf("%03d_%s.down.sql", version, safeName)

	return upFilename, downFilename
}

// CreateMigrationTemplate создает шаблон миграции
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

// MigrationInfo содержит информацию о миграции для отображения
type MigrationInfo struct {
	Version     int64  `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Applied     bool   `json:"applied"`
	AppliedAt   string `json:"applied_at,omitempty"`
	CanRollback bool   `json:"can_rollback"`
}

// GetMigrationInfo преобразует статус миграций в информацию для отображения
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
