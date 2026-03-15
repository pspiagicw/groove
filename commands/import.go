package commands

import (
	"fmt"
	"strings"

	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
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
			err := importItem(db, item)
			if err != nil {
				return fmt.Errorf("Error importing item: %v!", err)
			}
		}
	}

	return nil
}
func importItem(db *database.DB, item database.QueueInfo) error {

	// TODO: These functions will sanitize the data if needed.
	artists := getArtists(item)
	album := getAlbum(item)
	// albumArtist := getAlbumArtist(item)
	title := getTitle(item)

	albumID, err := db.InsertAlbum(album)
	if err != nil {
		return err
	}
	fmt.Println(albumID)

	trackID, err := db.InsertTrack(title, albumID)
	if err != nil {
		return err
	}
	fmt.Println(trackID)

	for _, artist := range artists {
		artistID, err := db.InsertArtist(artist)
		fmt.Println(artistID)
		if err != nil {
			return err
		}

		err = db.LinkTrackAndArtist(trackID, artistID)
		if err != nil {
			return err
		}
	}

	err = db.MarkProcessed(item)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Implement this functions properly.
func getArtists(item database.QueueInfo) []string {
	return strings.Split(item.DetectedArtist, ";")
}
func getAlbum(item database.QueueInfo) string {
	return item.DetectedAlbum
}
func getAlbumArtist(item database.QueueInfo) string {
	return item.DetectedAlbumArtist
}
func getTitle(item database.QueueInfo) string {
	return item.DetectedTitle
}
func confirmQueueItem(info database.QueueInfo) bool {
	fmt.Println("---- Import Queue Item ----")
	fmt.Printf("ID: %d\n", info.Id)
	fmt.Printf("Path: %s\n", info.Path)
	fmt.Printf("Hash: %s\n", info.Hash)
	fmt.Printf("Status: %s\n", info.Status)
	fmt.Printf("Detected Artist: %s\n", info.DetectedArtist)
	fmt.Printf("Detected Album Artist: %s\n", info.DetectedAlbumArtist)
	fmt.Printf("Detected Album: %s\n", info.DetectedAlbum)
	fmt.Printf("Detected Title: %s\n", info.DetectedTitle)

	fmt.Print("\nProceed? (y/n): ")

	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" {
		return true
	}

	return false
}
