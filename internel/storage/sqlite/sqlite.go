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
	"url-shortner/internel/domain/entities/redirectInfo"
	"url-shortner/internel/domain/entities/urlInfo"
	"url-shortner/internel/domain/entities/user"
	"url-shortner/internel/storage"
)

type Storage struct {
	Db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{Db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias string, userId float64) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.Db.Prepare("INSERT INTO url(url, alias, user_id) VALUES(?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		return 0, fmt.Errorf("%s, %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias, userId)
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

	stmt, err := s.Db.Prepare("SELECT url FROM url WHERE alias = ?")
	defer stmt.Close()
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

func (s *Storage) GetAllUrl(start, length int64) ([]urlInfo.UrlInfo, error) {
	const op = "storage.sqlite.GetAllUrl"
	query := `
		SELECT 
			u.alias, 
			u.url, 
			us.id, 
			us.username
		FROM 
			url u
		INNER JOIN 
			users us 
		ON 
			u.user_id = us.id 
		LIMIT ? OFFSET ?`
	stmt, err := s.Db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(length, start-1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var urls []urlInfo.UrlInfo

	for rows.Next() {
		var urlInfo urlInfo.UrlInfo
		var user user.User
		err := rows.Scan(&urlInfo.Alias, &urlInfo.Url, &user.ID, &user.Username)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		urlInfo.User = user
		urls = append(urls, urlInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return urls, nil
}

func (s *Storage) GetAllRedirectInfo(start, length int64) ([]redirectInfo.RedirectInfo, error) {
	const op = "storage.sqlite.GetAllRedirectInfo"
	query := `
		SELECT 
			id,
			ip,
			os,
			platform,
			browser,
			created_at
		FROM 
			url_redirection_info
		LIMIT ? OFFSET ?`
	stmt, err := s.Db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(length, start-1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var infos []redirectInfo.RedirectInfo

	for rows.Next() {
		var urlInfo redirectInfo.RedirectInfo
		err := rows.Scan(&urlInfo.Id, &urlInfo.Ip, &urlInfo.Os, &urlInfo.Platform, &urlInfo.Browser, &urlInfo.Created)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		infos = append(infos, urlInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return infos, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.Db.Prepare("DELETE FROM url WHERE alias = ?")
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	res, err := stmt.Exec(alias)
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

func (s *Storage) SaveRedirectInfo(redirectInfo *redirectInfo.RedirectInfo) error {
	const op = "storage.sqlite.SaveRedirectInfo"

	stmt, err := s.Db.Prepare("INSERT INTO url_redirection_info (ip, os, platform, browser) VALUES(?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	_, err = stmt.Exec(redirectInfo.Ip, redirectInfo.Os, redirectInfo.Platform, redirectInfo.Browser)

	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	return nil
}

func (s *Storage) GetUser(userName string) (user.User, error) {
	const op = "storage.sqlite.GetUser"

	stmt, err := s.Db.Prepare("SELECT id, username, password FROM users WHERE username = ?")
	defer stmt.Close()
	if err != nil {
		return user.User{}, fmt.Errorf("%s, %w", op, err)
	}

	var userEntity user.User
	err = stmt.QueryRow(userName).Scan(&userEntity.ID, &userEntity.Username, &userEntity.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, storage.UserNotFound
	}
	if err != nil {
		return user.User{}, fmt.Errorf("%s, %w", op, err)
	}

	return userEntity, nil
}

func (s *Storage) CloseConnection() {
	err := s.Db.Close()
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

		_, err = s.Db.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		log.Info("Executed migration: ", file)
	}

	return nil
}

func (s *Storage) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	const op = "storage.sqlite.Query"

	rows, err := s.Db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	results := make([]map[string]interface{}, 0)

	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnPointers {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("%s, %w", op, err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			rowMap[colName] = columnValues[i]
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	return results, nil
}
