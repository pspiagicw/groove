package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-homedir"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	dbPath, _ = homedir.Expand(dbPath)
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database: %v!", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v!", err)
	}

	db := &DB{
		conn: conn,
	}

	if err = db.initSchema(); err != nil {
		return nil, err
	}
	return db, nil
}

func (d *DB) Close() error {
	return d.Close()
}

func (d *DB) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS artists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS albums (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT,
	year INTEGER,
	album_artist_id INTEGER,
	FOREIGN KEY(album_artist_id) REFERENCES artists(id)
);

CREATE TABLE IF NOT EXISTS tracks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT,
	album_id INTEGER,
	track_number INTEGER,
	disc_number INTEGER,
	genre TEXT,
	duration INTEGER,
	FOREIGN KEY(album_id) REFERENCES albums(id)
);

CREATE TABLE IF NOT EXISTS track_artists (
	track_id INTEGER,
	artist_id INTEGER,
	role TEXT,
	PRIMARY KEY(track_id, artist_id),
	FOREIGN KEY(track_id) REFERENCES tracks(id),
	FOREIGN KEY(artist_id) REFERENCES artists(id)
);

CREATE TABLE IF NOT EXISTS files (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	track_id INTEGER,
	path TEXT UNIQUE,
	codec TEXT,
	bitrate INTEGER,
	sample_rate INTEGER,
	hash TEXT UNIQUE,
	size INTEGER,
	FOREIGN KEY(track_id) REFERENCES tracks(id)
);

CREATE TABLE IF NOT EXISTS import_queue (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT UNIQUE,
	hash TEXT,
	status TEXT,
	detected_artist TEXT,
	detected_title TEXT,
	detected_album TEXT
);
	`
	_, err := d.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
