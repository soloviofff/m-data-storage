package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

//go:embed sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed questdb/*.sql
var questdbMigrations embed.FS

// Manager управляет миграциями для обеих баз данных
type Manager struct {
	sqliteDB  *sql.DB
	questDB   *sql.DB
	logger    *logrus.Logger
	sqliteMgr *Migrator
	questMgr  *Migrator
}

// NewManager создает новый менеджер миграций
func NewManager(sqliteDB, questDB *sql.DB, logger *logrus.Logger) (*Manager, error) {
	if sqliteDB == nil {
		return nil, fmt.Errorf("SQLite database connection is required")
	}
	if questDB == nil {
		return nil, fmt.Errorf("QuestDB database connection is required")
	}
	if logger == nil {
		logger = logrus.New()
	}

	manager := &Manager{
		sqliteDB:  sqliteDB,
		questDB:   questDB,
		logger:    logger,
		sqliteMgr: NewMigrator(sqliteDB, "schema_migrations"),
		questMgr:  NewMigrator(questDB, "schema_migrations"),
	}

	// Загружаем миграции
	if err := manager.loadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return manager, nil
}

// loadMigrations загружает миграции из встроенных файлов
func (m *Manager) loadMigrations() error {
	// Загружаем SQLite миграции
	sqliteMigs, err := LoadMigrationsFromFS(sqliteMigrations, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to load SQLite migrations: %w", err)
	}

	if err := ValidateMigrations(sqliteMigs); err != nil {
		return fmt.Errorf("invalid SQLite migrations: %w", err)
	}

	m.sqliteMgr.AddMigrations(sqliteMigs)
	m.logger.WithField("count", len(sqliteMigs)).Info("Loaded SQLite migrations")

	// Загружаем QuestDB миграции
	questMigs, err := LoadMigrationsFromFS(questdbMigrations, "questdb")
	if err != nil {
		return fmt.Errorf("failed to load QuestDB migrations: %w", err)
	}

	if err := ValidateMigrations(questMigs); err != nil {
		return fmt.Errorf("invalid QuestDB migrations: %w", err)
	}

	m.questMgr.AddMigrations(questMigs)
	m.logger.WithField("count", len(questMigs)).Info("Loaded QuestDB migrations")

	return nil
}

// MigrateUp выполняет все неприменённые миграции для обеих баз данных
func (m *Manager) MigrateUp(ctx context.Context) error {
	m.logger.Info("Starting database migrations")
	start := time.Now()

	// Мигрируем SQLite
	m.logger.Info("Migrating SQLite database")
	if err := m.sqliteMgr.Up(ctx); err != nil {
		return fmt.Errorf("SQLite migration failed: %w", err)
	}
	m.logger.Info("SQLite migration completed successfully")

	// Мигрируем QuestDB
	m.logger.Info("Migrating QuestDB database")
	if err := m.questMgr.Up(ctx); err != nil {
		return fmt.Errorf("QuestDB migration failed: %w", err)
	}
	m.logger.Info("QuestDB migration completed successfully")

	duration := time.Since(start)
	m.logger.WithField("duration", duration).Info("All database migrations completed successfully")

	return nil
}

// MigrateDown откатывает миграции до указанной версии
func (m *Manager) MigrateDown(ctx context.Context, sqliteVersion, questdbVersion int64) error {
	m.logger.WithFields(logrus.Fields{
		"sqlite_version":  sqliteVersion,
		"questdb_version": questdbVersion,
	}).Info("Starting database rollback")

	// Откатываем QuestDB (сначала, так как данные зависят от метаданных)
	if questdbVersion >= 0 {
		m.logger.WithField("version", questdbVersion).Info("Rolling back QuestDB")
		if err := m.questMgr.Down(ctx, questdbVersion); err != nil {
			return fmt.Errorf("QuestDB rollback failed: %w", err)
		}
		m.logger.Info("QuestDB rollback completed")
	}

	// Откатываем SQLite
	if sqliteVersion >= 0 {
		m.logger.WithField("version", sqliteVersion).Info("Rolling back SQLite")
		if err := m.sqliteMgr.Down(ctx, sqliteVersion); err != nil {
			return fmt.Errorf("SQLite rollback failed: %w", err)
		}
		m.logger.Info("SQLite rollback completed")
	}

	m.logger.Info("Database rollback completed successfully")
	return nil
}

// Status возвращает статус миграций для обеих баз данных
func (m *Manager) Status(ctx context.Context) (*DatabaseMigrationStatus, error) {
	// Получаем статус SQLite
	sqliteStatus, err := m.sqliteMgr.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQLite migration status: %w", err)
	}

	// Получаем статус QuestDB
	questdbStatus, err := m.questMgr.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get QuestDB migration status: %w", err)
	}

	return &DatabaseMigrationStatus{
		SQLite:  GetMigrationInfo(sqliteStatus),
		QuestDB: GetMigrationInfo(questdbStatus),
	}, nil
}

// DatabaseMigrationStatus представляет общий статус миграций
type DatabaseMigrationStatus struct {
	SQLite  []MigrationInfo `json:"sqlite"`
	QuestDB []MigrationInfo `json:"questdb"`
}

// ValidateConnections проверяет подключения к базам данных
func (m *Manager) ValidateConnections(ctx context.Context) error {
	// Проверяем SQLite
	if err := m.sqliteDB.PingContext(ctx); err != nil {
		return fmt.Errorf("SQLite connection failed: %w", err)
	}

	// Проверяем QuestDB
	if err := m.questDB.PingContext(ctx); err != nil {
		return fmt.Errorf("QuestDB connection failed: %w", err)
	}

	m.logger.Info("Database connections validated successfully")
	return nil
}

// GetSQLiteMigrator возвращает мигратор для SQLite
func (m *Manager) GetSQLiteMigrator() *Migrator {
	return m.sqliteMgr
}

// GetQuestDBMigrator возвращает мигратор для QuestDB
func (m *Manager) GetQuestDBMigrator() *Migrator {
	return m.questMgr
}

// Close закрывает соединения с базами данных
func (m *Manager) Close() error {
	var sqliteErr, questErr error

	if m.sqliteDB != nil {
		sqliteErr = m.sqliteDB.Close()
	}

	if m.questDB != nil {
		questErr = m.questDB.Close()
	}

	if sqliteErr != nil {
		return fmt.Errorf("failed to close SQLite connection: %w", sqliteErr)
	}

	if questErr != nil {
		return fmt.Errorf("failed to close QuestDB connection: %w", questErr)
	}

	return nil
}
