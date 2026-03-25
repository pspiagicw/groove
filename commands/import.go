package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
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
			importItem(conf, db, session)
			// prettylog.Infof("Item imported!")
		}
	}

	return nil
}

func extractMainArtist(name string) string {
	splits := strings.Split(name, ",")
	return splits[0]
}
func importItem(conf *config.Config, db *database.DB, item *ImportSession) error {

	// TODO: Add genre, year, disc, track_number and album artist id.
	album := item.NormalizedAlbum
	title := item.NormalizedTitle
	artists := item.NormalizedArtists
	album_artist := item.NormalizedAlbumArtist
	year := item.NormalizedYear
	track_number := item.NormalizedTrackNumber
	disc := item.NormalizedDiscNumber

	album_artist = extractMainArtist(album_artist)

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

	// err = moveFile(conf, item)
	// if err != nil {
	// 	return fmt.Errorf("Error moving file: %v!", err)
	// }
	//
	// err = db.MarkProcessed(item)
	// if err != nil {
	// 	return err
	// }

	// prettylog.Successf("Imported %q", item.DetectedTitle)

	return nil
}
func moveFile(conf *config.Config, item *ImportSession) error {
	err := utils.CreateIfNotExist(conf.LibraryDir)
	if err != nil {
		return fmt.Errorf("Error creating library folder: %v!", err)
	}
	extension := filepath.Ext(item.Path)

	newPath := filepath.Join(conf.LibraryDir, item.NormalizedAlbum, fmt.Sprintf("%s · %s%s", item.NormalizedTitle, item.DetectedArtist, extension))
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

func getItemDetails(info *database.ScannedItem) *ImportSession {

	item := confirmItemDetails(info)

	if item.Skipped {
		prettylog.Infof("(%s) skipped!", info.Path)
		return nil
	}

	// fmt.Println(item)

	return &item
}
