package database

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type URL struct {
	ID        int64
	Code      string
	LongURL   string
	CreatedAt time.Time
	Clicks    int64
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS urls (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    code      TEXT NOT NULL UNIQUE,
    long_url  TEXT NOT NULL,
    clicks    INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_urls_code ON urls(code);
`)
	return err
}

func (s *Store) Insert(code, longURL string) error {
	_, err := s.db.Exec(`INSERT INTO urls(code, long_url) VALUES(?, ?)`, code, longURL)
	return err
}

func (s *Store) GetByCode(code string) (*URL, error) {
	var u URL
	err := s.db.QueryRow(`
SELECT id, code, long_url, created_at, clicks
FROM urls WHERE code = ?`, code).Scan(&u.ID, &u.Code, &u.LongURL, &u.CreatedAt, &u.Clicks)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	return &u, err
}

func (s *Store) IncrementClick(id int64) error {
	_, err := s.db.Exec(`UPDATE urls SET clicks = clicks + 1 WHERE id = ?`, id)
	return err
}
