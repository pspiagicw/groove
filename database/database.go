package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DBTX interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}
type DB struct {
	conn *sql.DB
}

type QueueItem struct {
	Id                  int
	Path                string
	Hash                string
	Status              string
	DetectedArtist      string
	DetectedAlbum       string
	DetectedTitle       string
	DetectedAlbumArtist string
	DetectedYear        int
	DetectedGenre       string
}

func NewDB(dbPath string) (*DB, error) {
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

func (d *DB) QueryQueue() ([]QueueItem, error) {
	queryResult := []QueueItem{}
	rows, err := d.conn.Query("SELECT * FROM import_queue where status = 'pending';")
	if err != nil {
		return nil, fmt.Errorf("Error querying for import queue: %v!", err)
	}

	for rows.Next() {
		info := new(QueueItem)
		err := rows.Scan(&info.Id, &info.Path, &info.Hash, &info.Status, &info.DetectedArtist, &info.DetectedTitle, &info.DetectedAlbum, &info.DetectedAlbumArtist)
		if err != nil {
			return nil, fmt.Errorf("Error scanning query into struct: %v!", err)
		}

		queryResult = append(queryResult, *info)
	}

	return queryResult, nil
}

func (d *DB) InsertArtist(artist string) (int, error) {
	_, err := d.conn.Exec("INSERT INTO artists(name) VALUES (?) ON CONFLICT(NAME) DO NOTHING;", artist)
	if err != nil {
		return -1, fmt.Errorf("Error inserting artist: %v!", err)
	}

	row := d.conn.QueryRow("SELECT id from artists where name = ?;", artist)

	var id int
	err = row.Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("Error finding id of inserted artist: %v!", err)
	}

	return id, nil
}

func (d *DB) InsertAlbum(album string) (int, error) {

	_, err := d.conn.Exec("INSERT INTO albums(title) values (?) on conflict(title) do nothing;", album)
	if err != nil {
		return -1, fmt.Errorf("Error inserting album: %v!", err)
	}

	row := d.conn.QueryRow("SELECT id from albums where title = ?;", album)

	var id int
	err = row.Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("Error finding id of inserted album: %v!", err)
	}

	return id, nil
}

func (d *DB) InsertTrack(title string, albumID int) (int, error) {

	_, err := d.conn.Exec("INSERT INTO tracks(title, album_id) values (?, ?) on conflict(title) do nothing;", title, albumID)
	if err != nil {
		return -1, fmt.Errorf("Error inserting track: %v!", err)
	}

	row := d.conn.QueryRow("SELECT id from tracks where title = ?;", title)

	var id int
	err = row.Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("Error finding id of inserted track: %v!", err)
	}

	return id, nil
}

func (d *DB) LinkTrackAndArtist(trackID int, artistID int) error {
	_, err := d.conn.Exec("INSERT INTO track_artists(track_id, artist_id) values (?, ?);", trackID, artistID)

	if err != nil {
		return fmt.Errorf("Error linking track with artist: %v!", err)
	}
	return nil
}

func (d *DB) AddToQueue(queueInfo []QueueItem) (int, error) {
	rowsAffected := 0
	stmt, err := d.conn.Prepare("INSERT OR IGNORE INTO import_queue(path, hash, status, detected_artist, detected_title, detected_album, detected_album_artist) values (?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return 0, fmt.Errorf("Error while inserting into db: %v!", err)
	}

	for _, info := range queueInfo {
		result, err := stmt.Exec(info.Path, info.Hash, info.Status, info.DetectedArtist, info.DetectedTitle, info.DetectedAlbum, info.DetectedAlbumArtist)
		if err != nil {
			return 0, fmt.Errorf("Error while inserting into db: %v!", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("Error confirming insertion: %v", err)
		}
		rowsAffected += int(rows)
	}

	err = stmt.Close()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (d *DB) MarkProcessed(queueInfo QueueItem) error {

	_, err := d.conn.Exec("UPDATE import_queue SET status = 'imported' where id = ?", queueInfo.Id)

	if err != nil {
		return fmt.Errorf("Error marking queue item: %v!", err)
	}

	return nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS artists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS albums (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT UNIQUE,
	year INTEGER,
	album_artist_id INTEGER,
	FOREIGN KEY(album_artist_id) REFERENCES artists(id)
);

CREATE TABLE IF NOT EXISTS tracks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT UNIQUE,
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
	hash TEXT UNIQUE,
	status TEXT,
	detected_artist TEXT,
	detected_title TEXT,
	detected_album TEXT,
	detected_album_artist TEXT
);
	`
	_, err := d.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
