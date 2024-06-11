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
	Id       int64  `json:"id,omitempty"`
	Ip       string `json:"ip"`
	Os       string `json:"os"`
	Platform string `json:"platform"`
	Browser  string `json:"browser"`
	Created  string `json:"created"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

type UrlInfo struct {
	Alias string `json:"alias"`
	Url   string `json:"url"`
	User  User   `json:"user"`
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias string, userId float64) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias, user_id) VALUES(?, ?, ?)")
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

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
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

func (s *Storage) GetAllUrl(start, length int64) ([]UrlInfo, error) {
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
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(length, start-1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var urls []UrlInfo

	for rows.Next() {
		var urlInfo UrlInfo
		var user User
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

func (s *Storage) GetAllRedirectInfo(start, length int64) ([]RedirectInfo, error) {
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
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(length, start-1)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var infos []RedirectInfo

	for rows.Next() {
		var urlInfo RedirectInfo
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

func (s *Storage) DeleteURL(id int64) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE id = ?")
	defer stmt.Close()
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

func (s *Storage) GetUser(userName string) (User, error) {
	const op = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare("SELECT id, username, password FROM users WHERE username = ?")
	defer stmt.Close()
	if err != nil {
		return User{}, fmt.Errorf("%s, %w", op, err)
	}

	var user User
	err = stmt.QueryRow(userName).Scan(&user.ID, &user.Username, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, storage.UserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("%s, %w", op, err)
	}

	return user, nil
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

func (s *Storage) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	const op = "storage.sqlite.Query"

	rows, err := s.db.Query(query, args...)
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
