package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/excircle/quik-version/internal/config"
)

const schema = `
CREATE TABLE IF NOT EXISTS versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    git_sha TEXT NOT NULL,
    git_url TEXT NOT NULL,
    increment_type TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(git_url, version)
);

CREATE TABLE IF NOT EXISTS config_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_synced_at TIMESTAMP,
    git_url TEXT NOT NULL
);
`

// DB wraps the SQLite database connection
type DB struct {
	*sql.DB
	path string
}

// GetDBPath returns the path where qv.db should be stored
func GetDBPath() string {
	dbPath := config.GetDBPath()
	if dbPath != "" {
		return filepath.Join(dbPath, "qv.db")
	}
	return "qv.db"
}

// Exists checks if the database file exists
func Exists() bool {
	path := GetDBPath()
	_, err := os.Stat(path)
	return err == nil
}

// Open opens or creates the database
func Open() (*DB, error) {
	path := GetDBPath()

	// Ensure directory exists if custom path specified
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &DB{DB: db, path: path}, nil
}

// Initialize creates the database schema
func (db *DB) Initialize() error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

// SetConfigState sets or updates the config state
func (db *DB) SetConfigState(gitURL string) error {
	_, err := db.Exec(`
		INSERT INTO config_state (id, git_url, last_synced_at)
		VALUES (1, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET git_url = ?, last_synced_at = CURRENT_TIMESTAMP
	`, gitURL, gitURL)
	if err != nil {
		return fmt.Errorf("failed to set config state: %w", err)
	}
	return nil
}

// GetLatestVersion returns the latest version record for a git URL
func (db *DB) GetLatestVersion(gitURL string) (*Version, error) {
	row := db.QueryRow(`
		SELECT id, version, tag_name, git_sha, git_url, increment_type, created_at
		FROM versions
		WHERE git_url = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, gitURL)

	var v Version
	err := row.Scan(&v.ID, &v.Version, &v.TagName, &v.GitSHA, &v.GitURL, &v.IncrementType, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	return &v, nil
}

// InsertVersion adds a new version record
func (db *DB) InsertVersion(v *Version) error {
	_, err := db.Exec(`
		INSERT INTO versions (version, tag_name, git_sha, git_url, increment_type)
		VALUES (?, ?, ?, ?, ?)
	`, v.Version, v.TagName, v.GitSHA, v.GitURL, v.IncrementType)
	if err != nil {
		return fmt.Errorf("failed to insert version: %w", err)
	}
	return nil
}

// GetAllVersions returns all versions for a git URL
func (db *DB) GetAllVersions(gitURL string) ([]Version, error) {
	rows, err := db.Query(`
		SELECT id, version, tag_name, git_sha, git_url, increment_type, created_at
		FROM versions
		WHERE git_url = ?
		ORDER BY created_at DESC
	`, gitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []Version
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.ID, &v.Version, &v.TagName, &v.GitSHA, &v.GitURL, &v.IncrementType, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// Version represents a version record
type Version struct {
	ID            int
	Version       string
	TagName       string
	GitSHA        string
	GitURL        string
	IncrementType *string
	CreatedAt     string
}
