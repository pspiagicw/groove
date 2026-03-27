package commands

import (
	"path/filepath"

	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
	"github.com/pspiagicw/groove/prettylog"
	"github.com/pspiagicw/groove/utils"
)

const SINGLES_FOLDER = "Singles"

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
			moveToFolder(config, file.Path, SINGLES_FOLDER, title)
		} else {
			moveToFolder(config, file.Path, albumName, title)
		}
	}

	return nil
}
func moveToFolder(config *config.Config, origPath string, foldername string, title string) {
	library := config.LibraryDir
	ext := filepath.Ext(origPath)

	parentFolder := "Albums"

	if foldername == SINGLES_FOLDER {
		parentFolder = ""
	}

	newPath := filepath.Join(library, parentFolder, foldername, title+ext)

	prettylog.Infof("Copying (%s) to %s", origPath, newPath)

	utils.CreateIfNotExist(filepath.Dir(newPath))

	err := utils.CopyFile(origPath, newPath)

	if err != nil {
		prettylog.Errorf("Error moving file(%s): %v!", origPath, err)
	}
}
