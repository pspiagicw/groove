package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
)

func Scan(configPath string) error {
	conf, err := config.ConfigProvider(configPath)
	if err != nil {
		return fmt.Errorf("Error loading config: %v", err)
	}

	db, err := database.NewDB(conf.Database)
	if err != nil {
		return fmt.Errorf("Error connecting to database: %v", err)
	}

	fmt.Println(conf.IncomingDir)
	files, err := globFiles(conf.IncomingDir)

	queueInfo, err := processFiles(files)
	// fmt.Println(queueInfo)
	rowsAffected, err := db.AddToQueue(queueInfo)
	if err != nil {
		return fmt.Errorf("Error inserting queue: %v!", err)
	}
	fmt.Printf("Rows affected: %d\n", rowsAffected)
	// TODO: Add n files duplicate thing.

	err = db.Close()
	if err != nil {
		return fmt.Errorf("Error closing database: %v!", err)
	}

	return nil
}

func globFiles(incomingDir string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(incomingDir, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() {
			// If directory, skip for now (TODO: Add album feature maybe!)
			return nil
		}

		if isMusicFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}
func isMusicFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".mp3" || ext == ".flac" || ext == ".opus" || ext == ".wav" || ext == ".m4a" || ext == ".ogg"
}
func processFiles(files []string) ([]database.QueueInfo, error) {
	queueInfo := []database.QueueInfo{}
	for _, filepath := range files {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("Error opening file: %v!", err)
		}

		metadata, err := tag.ReadFrom(f)
		if err != nil {
			return nil, fmt.Errorf("Error reading metadata: %v!", err)
		}

		hash, err := tag.Sum(f)
		if err != nil {
			return nil, fmt.Errorf("Error calculating hash: %v!", err)
		}

		info := database.QueueInfo{
			Path:           filepath,
			Hash:           hash,
			Status:         "pending",
			DetectedArtist: metadata.Artist(),
			DetectedTitle:  metadata.Title(),
			DetectedAlbum:  metadata.Album(),
		}

		queueInfo = append(queueInfo, info)
	}

	return queueInfo, nil
}
