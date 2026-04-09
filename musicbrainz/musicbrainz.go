package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Release struct {
	Title string `json:"title"`
	Date  string `json:"date"`
	ID    string `json: "id"`
}

type Recording struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Length  int    `json:"length"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artist-credit"`

	Releases []Release `json:"releases"`
}
type MBResponse struct {
	Recordings []Recording `json:"recordings"`
}

func Query(title string, artist string) ([]Recording, error) {
	query := url.QueryEscape(fmt.Sprintf(`recording: "%s" AND artist:"%s"`, title, artist))

	url := fmt.Sprintf(
		"https://musicbrainz.org/ws/2/recording?query=%s&fmt=json&inc=artists+releases",
		query,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"User-Agent",
		"groove/0.1 ( https://github.com/pspiagicw/groove )",
	)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result MBResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if len(result.Recordings) == 0 {
		return nil, fmt.Errorf("no results")
	}

	return result.Recordings, nil

}

// func LookupMusicBrainz(artist string, title string) (*Recording, error) {
//
// 	query := url.QueryEscape(fmt.Sprintf(`recording:"%s" AND artist:"%s"`, title, artist))
//
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Set(
// 		"User-Agent",
// 		"groove/0.1 ( https://github.com/yourname/groove )",
// 	)
//
// 	client := &http.Client{
// 		Timeout: 10 * time.Second,
// 	}
//
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
//
// 	var result MBResponse
//
// 	err = json.NewDecoder(resp.Body).Decode(&result)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if len(result.Recordings) == 0 {
// 		return nil, fmt.Errorf("no results")
// 	}
//
// 	return &result.Recordings[0], nil
// }
