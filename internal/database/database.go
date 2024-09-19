package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db  *pgxpool.Pool
	dsn string
}

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}
//go:embed migrations/*
var migrationsFS embed.FS // Встраивание файлов миграций

// checkConnection проверяет доступность базы данных с использованием пулa подключений.
func checkConnection(ctx context.Context, db *pgxpool.Pool) error {
	// Устанавливаем таймаут для проверки подключения
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Пингуем базу данных для проверки доступности
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	return nil
}

// New создает новый экземпляр Database, устанавливает соединение и проверяет его.
func New(ctx context.Context, dsn string) (*Database, error) {
	// Создаем пул подключений к базе данных
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пула подключений: %w", err)
	}

	// Проверяем доступность базы данных
	if err := checkConnection(ctx, db); err != nil {
		db.Close() // Закрываем пул подключений в случае ошибки
		return nil, err
	}

	return &Database{db: db, dsn: dsn}, nil
}

// RunMigrations выполняет миграции базы данных с использованием встроенных файлов миграций.
func (d *Database) RunMigrations() error {
	// Создаем источник миграций из встроенных файлов
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("не удалось создать источник миграций: %w", err)
	}

	// Инициализируем миграции с использованием источника и строки подключения
	migrations, err := migrate.NewWithSourceInstance("iofs", driver, d.dsn)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать миграции: %w", err)
	}

	// Применяем миграции
	err = migrations.Up()
	if err != nil {
		// Обрабатываем ошибку отсутствия новых миграций отдельно
		if err == migrate.ErrNoChange {
			log.Println("Новых миграций не найдено")
			return nil
		}
		return fmt.Errorf("ошибка при выполнении миграций: %w", err)
	}

	log.Println("Миграции успешно применены")
	return nil
}

// Close закрывает пул подключений к базе данных.
func (d *Database) Close() {
	if d.db != nil {
		d.db.Close()
	}
}