package commands

import (
	"fmt"
	"path/filepath"

	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
	"github.com/pspiagicw/groove/prettylog"
)

func Move(configPath string) error {
	config := config.ConfigProvider(configPath)

	db := database.NewDB(config.Database)

	files := db.QueryFiles()

	for _, file := range files {
		albumName, trackCount, err := db.QueryAlbumAndTrackCount(file.TrackID)

		if err != nil {
			prettylog.Fatalf("Failed to find information in db: %v!", err)
		}

		title, err := db.QueryTitle(file.TrackID)

		if err != nil {
			prettylog.Fatalf("Failed to find title for given trackID: %v!", err)
		}

		// TODO: Make this parameterzied
		if trackCount < 3 {
			moveToFolder(config, file.Path, "Single", title)
		} else {
			moveToFolder(config, file.Path, albumName, title)
		}
	}

	return nil
}
func moveToFolder(config *config.Config, origPath string, foldername string, title string) {
	library := config.LibraryDir
	ext := filepath.Ext(origPath)

	newPath := filepath.Join(library, foldername, title+ext)

	fmt.Println(newPath)
}
