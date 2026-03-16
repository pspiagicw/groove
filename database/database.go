package database

import (
	"database/sql"
	"fmt"

	"github.com/dhowden/tag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pspiagicw/groove/prettylog"
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

type ScannedItem struct {
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
	DetectedDisc        int
	DetectedTrackNumber int
	DetectedFileType    tag.FileType
}

func NewDB(dbPath string) *DB {
	conn, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		prettylog.Fatalf("Failed to open database connection: %v!", err)
	}

	if err = conn.Ping(); err != nil {
		prettylog.Fatalf("Failed to ping database: %v", err)
	}

	db := &DB{
		conn: conn,
	}

	db.initSchema()

	return db
}

func (d *DB) QueryQueue() []ScannedItem {

	queryResult := []ScannedItem{}
	rows, err := d.conn.Query("SELECT * FROM import_queue where status = 'pending';")

	if err != nil {
		prettylog.Fatalf("Failed to query items from database: %v!", err)
	}

	for rows.Next() {

		info := new(ScannedItem)
		err := rows.Scan(
			&info.Id,
			&info.Path,
			&info.Hash,
			&info.Status,
			&info.DetectedArtist,
			&info.DetectedTitle,
			&info.DetectedAlbum,
			&info.DetectedAlbumArtist,
			&info.DetectedYear,
			&info.DetectedGenre,
			&info.DetectedDisc,
			&info.DetectedTrackNumber,
			&info.DetectedFileType,
		)

		if err != nil {
			prettylog.Fatalf("Failed to scan sql result into struct: %v!", err)
		}

		queryResult = append(queryResult, *info)
	}

	return queryResult
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

func (d *DB) AddToQueue(queueInfo []ScannedItem) int {
	rowsAffected := 0
	stmt, err := d.conn.Prepare("INSERT OR IGNORE INTO import_queue(path, hash, status, detected_artist, detected_title, detected_album, detected_album_artist, detected_year, detected_genre, detected_disc, detected_track_number, detected_filetype) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		prettylog.Fatalf("Failed to insert item into queue: %v!", err)
	}

	for _, info := range queueInfo {
		result, err := stmt.Exec(info.Path, info.Hash, info.Status, info.DetectedArtist, info.DetectedTitle, info.DetectedAlbum, info.DetectedAlbumArtist, info.DetectedYear, info.DetectedGenre, info.DetectedDisc, info.DetectedTrackNumber, info.DetectedFileType)

		if err != nil {
			prettylog.Fatalf("Failed to insert item into db: %v!", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			prettylog.Fatalf("Failed to fetch item processed: %v", err)
		}
		rowsAffected += int(rows)
	}

	err = stmt.Close()
	if err != nil {
		prettylog.Fatalf("Failed to close database statement: %v", err)
	}

	return rowsAffected
}

func (d *DB) MarkProcessed(queueInfo ScannedItem) error {

	_, err := d.conn.Exec("UPDATE import_queue SET status = 'imported' where id = ?", queueInfo.Id)

	if err != nil {
		return fmt.Errorf("Error marking queue item: %v!", err)
	}

	return nil
}

func (d *DB) Close() {
	err := d.conn.Close()
	if err != nil {
		prettylog.Fatalf("Failed to close database: %v!", err)
	}
}

func (d *DB) initSchema() {
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
	detected_album_artist TEXT,
	detected_year INTEGER,
	detected_genre TEXT,
	detected_disc INTEGER,
	detected_track_number INTEGER,
	detected_filetype TEXT
);
	`
	_, err := d.conn.Exec(schema)
	if err != nil {
		prettylog.Fatalf("Failed to initialize schema: %v", err)
	}
}
