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
	conf := config.ConfigProvider(configPath)

	db := database.NewDB(conf.Database)

	queue := db.QueryQueue()

	for _, item := range queue {

		session := getItemDetails(&item)

		if session != nil {
			// importItem(conf, db, item)
			prettylog.Infof("Item imported!")
		}
	}

	return nil
}
func importItem(conf *config.Config, db *database.DB, item database.ScannedItem) error {

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
func moveFile(conf *config.Config, item database.ScannedItem) error {
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
func getArtists(item database.ScannedItem) []string {
	return strings.Split(item.DetectedArtist, ";")
}
func getAlbum(item database.ScannedItem) string {
	return item.DetectedAlbum
}
func getAlbumArtist(item database.ScannedItem) string {
	return item.DetectedAlbumArtist
}
func getTitle(item database.ScannedItem) string {
	return item.DetectedTitle
}

func formatArtists(recording musicbrainz.Recording) string {
	artists := make([]string, 0, len(recording.Artists))
	for _, artist := range recording.Artists {
		artists = append(artists, artist.Name)
	}

	return strings.Join(artists, ", ")
}

func printRecording(recording musicbrainz.Recording) {
	prettylog.PrintBlock(
		os.Stdout,
		"MusicBrainz Match",
		prettylog.KV("Title", recording.Title),
		prettylog.KV("Artists", formatArtists(recording)),
		prettylog.KV("Release Count", len(recording.Releases)),
		prettylog.KV("Length (ms)", recording.Length),
	)
}

func getItemDetails(info *database.ScannedItem) *ImportSession {

	item := confirmItemDetails(info)

	if item.Skipped {
		prettylog.Infof("(%s) skipped!", info.Path)
		return nil
	}

	fmt.Println(item)

	return &item
}
