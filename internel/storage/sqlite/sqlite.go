package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"modernc.org/sqlite"
	_ "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
	"os"
	"path/filepath"
	"url-shortner/internel/storage"
)

type Storage struct {
	db *sql.DB
}

type RedirectInfo struct {
	Ip       string
	Os       string
	Platform string
	Browser  string
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s, %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var liteErr *sqlite.Error
		if errors.As(err, &liteErr) {
			if liteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return 0, storage.ErrURLExists
			}
		}
		return 0, fmt.Errorf("%s, %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s, %w", op, err)
	}

	var url string
	err = stmt.QueryRow(alias).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}

	return url, nil
}

func (s *Storage) DeleteURL(id int64) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	} else if affectedRows == 0 {
		return storage.ErrIdNotFound
	}

	return nil
}

func (s *Storage) SaveRedirectInfo(redirectInfo *RedirectInfo) error {
	const op = "storage.sqlite.SaveRedirectInfo"

	stmt, err := s.db.Prepare("INSERT INTO url_redirection_info (ip, os, platform, browser) VALUES(?, ?, ?, ?)")

	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	_, err = stmt.Exec(redirectInfo.Ip, redirectInfo.Os, redirectInfo.Platform, redirectInfo.Browser)

	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	return nil
}

func (s *Storage) CloseConnection() {
	err := s.db.Close()
	if err != nil {
		panic(err)
	}
}

func (s *Storage) RunMigrations(log *slog.Logger) error {
	const op = "storage.sqlite.initMigrations"

	migrationDir := filepath.Join("storage", "migrations")

	migrationFiles, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, file := range migrationFiles {
		migration, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		_, err = s.db.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		log.Info("Executed migration: ", file)
	}

	return nil
}
