package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
	"github.com/pspiagicw/groove/musicbrainz"
	"github.com/pspiagicw/groove/prettylog"
	"github.com/pspiagicw/groove/utils"
)

func Import(configPath string) error {
	conf, err := config.ConfigProvider(configPath)
	if err != nil {
		return fmt.Errorf("Error loading config: %v!", err)
	}

	db, err := database.NewDB(conf.Database)

	queue, err := db.QueryQueue()
	if err != nil {
		return fmt.Errorf("Error querying queue: %v!", err)
	}

	for _, item := range queue {
		if confirmQueueItem(item) {
			err := importItem(conf, db, item)
			if err != nil {
				return fmt.Errorf("Error importing item: %v!", err)
			}
		}
	}

	return nil
}
func importItem(conf *config.Config, db *database.DB, item database.QueueItem) error {

	// TODO: These functions will sanitize the data if needed.
	artists := getArtists(item)
	album := getAlbum(item)
	// albumArtist := getAlbumArtist(item)
	title := getTitle(item)

	albumID, err := db.InsertAlbum(album)
	if err != nil {
		return err
	}
	prettylog.Infof("Linked album %q as id=%d", album, albumID)

	trackID, err := db.InsertTrack(title, albumID)
	if err != nil {
		return err
	}
	prettylog.Infof("Linked track %q as id=%d", title, trackID)

	for _, artist := range artists {
		artistID, err := db.InsertArtist(artist)
		if err != nil {
			return err
		}
		prettylog.Infof("Linked artist %q as id=%d", artist, artistID)

		err = db.LinkTrackAndArtist(trackID, artistID)
		if err != nil {
			return err
		}
	}

	err = moveFile(conf, item)
	if err != nil {
		return fmt.Errorf("Error moving file: %v!", err)
	}

	err = db.MarkProcessed(item)
	if err != nil {
		return err
	}

	prettylog.Successf("Imported %q", item.DetectedTitle)

	return nil
}
func moveFile(conf *config.Config, item database.QueueItem) error {
	err := utils.CreateIfNotExist(conf.LibraryDir)
	if err != nil {
		return fmt.Errorf("Error creating library folder: %v!", err)
	}
	extension := filepath.Ext(item.Path)

	newPath := filepath.Join(conf.LibraryDir, item.DetectedAlbum, fmt.Sprintf("%s · %s%s", item.DetectedTitle, item.DetectedArtist, extension))
	err = utils.CreateIfNotExist(filepath.Dir(newPath))
	if err != nil {
		return fmt.Errorf("Error creating directory for file: %v", err)
	}

	if utils.AlreadyExists(newPath) {
		return fmt.Errorf("Error copying file, destination file already exists!")
	}

	err = utils.CopyFile(item.Path, newPath)
	if err != nil {
		return fmt.Errorf("Error copying file: %v!", err)
	}

	return nil
}

// TODO: Implement this functions properly.
func getArtists(item database.QueueItem) []string {
	return strings.Split(item.DetectedArtist, ";")
}
func getAlbum(item database.QueueItem) string {
	return item.DetectedAlbum
}
func getAlbumArtist(item database.QueueItem) string {
	return item.DetectedAlbumArtist
}
func getTitle(item database.QueueItem) string {
	return item.DetectedTitle
}

func formatArtists(recording *musicbrainz.Recording) string {
	artists := make([]string, 0, len(recording.Artists))
	for _, artist := range recording.Artists {
		artists = append(artists, artist.Name)
	}

	return strings.Join(artists, ", ")
}

func confirmQueueItem(info database.QueueItem) bool {
	prettylog.PrintBlock(
		os.Stdout,
		"Import Queue Item",
		prettylog.KV("ID", info.Id),
		prettylog.KV("Path", info.Path),
		prettylog.KV("Hash", info.Hash),
		prettylog.KV("Status", info.Status),
		prettylog.KV("Artist", info.DetectedArtist),
		prettylog.KV("Album Artist", info.DetectedAlbumArtist),
		prettylog.KV("Album", info.DetectedAlbum),
		prettylog.KV("Title", info.DetectedTitle),
	)

	recording, err := musicbrainz.Query(info.DetectedTitle, info.DetectedArtist)
	if err != nil {
		prettylog.Fatalf("MusicBrainz query failed: %v", err)
	}
	prettylog.PrintBlock(
		os.Stdout,
		"MusicBrainz Match",
		prettylog.KV("Title", recording.Title),
		prettylog.KV("Artists", formatArtists(recording)),
		prettylog.KV("Release Count", len(recording.Releases)),
		prettylog.KV("Length (ms)", recording.Length),
	)

	fmt.Print(prettylog.Prompt("Proceed? (y/n): "))

	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" {
		return true
	}

	return false
}
