package commands

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
	"github.com/pspiagicw/groove/prettylog"
	"github.com/pspiagicw/groove/utils"
)

func Scan(configPath string) {

	conf := config.ConfigProvider(configPath)

	db := database.NewDB(conf.Database)

	prettylog.Infof("Scanning %s", conf.IncomingDir)
	files := globFiles(conf.IncomingDir)

	prettylog.Infof("Scanned %d files.", len(files))

	queueInfo := processFiles(files)

	prettylog.Infof("Processed %d files.", len(files))

	rowsAffected, err := db.AddToQueue(queueInfo)

	if err != nil {
		prettylog.Fatalf("Error inserting into queue: %v!", err)
	}

	// DONE: Add n files duplicate thing.
	prettylog.Successf("%d items inserted into queue. found %d duplicates!", rowsAffected, len(queueInfo))
	prettylog.Successf("Queued %d file(s) for import", rowsAffected)

	err = db.Close()
	if err != nil {
		prettylog.Fatalf("Failed to close database: %v!", err)
	}
}

func globFiles(incomingDir string) []string {
	files := []string{}
	err := filepath.WalkDir(incomingDir, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			prettylog.Errorf("Error while recursing directory: %v!", err)
			return err
		}

		if dirEntry.IsDir() {
			// If directory, skip for now (TODO: Add album feature maybe!)
			return nil
		}

		if utils.IsMusicFile(path) {
			files = append(files, path)
		}

		return nil
	})

	// We only log a error and not fatalf, cause we can still import whatever was scanned.
	if err != nil {
		prettylog.Errorf("Error while recursing directory: %v!", err)
	}
	return files
}

func processFiles(files []string) []database.QueueItem {
	queueInfo := []database.QueueItem{}
	// We try to process as many files as possible.
	// Thus Errorf and not Fatalf
	for _, filepath := range files {
		f, err := os.Open(filepath)
		if err != nil {
			// We again try to process as many files as we can.
			prettylog.Errorf("Failed to open file(%s): %v!", filepath, err)
			continue
		}

		metadata, err := tag.ReadFrom(f)
		if err != nil {
			prettylog.Errorf("Failed to read metadata(%s): %v!", filepath, err)
			continue
		}

		hash, err := tag.Sum(f)
		if err != nil {
			prettylog.Errorf("Failed to calculate checksum(%s): %v!", filepath, err)
			continue
		}

		info := database.QueueItem{
			Path:           filepath,
			Hash:           hash,
			Status:         "pending",
			DetectedArtist: metadata.Artist(),
			DetectedTitle:  metadata.Title(),
			DetectedAlbum:  metadata.Album(),
			DetectedYear:   metadata.Year(),
			DetectedGenre:  metadata.Genre(),
		}

		queueInfo = append(queueInfo, info)
	}

	return queueInfo
}
