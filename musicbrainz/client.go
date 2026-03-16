package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type EnrichedRecording struct {
	RecordingID string
	ReleaseID   string

	Title   string
	Artists []string
	Length  int

	Album        string
	AlbumArtists []string
	Year         int
	Genres       []string

	TrackNumber int
	DiscNumber  int
}

type mbClient struct {
	httpClient *http.Client
	userAgent  string
	limiter    <-chan time.Time
}

func NewMBClient(userAgent string) *mbClient {
	return &mbClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		userAgent:  userAgent,
		limiter:    time.Tick(1 * time.Second),
	}
}

type releaseLookupResponse struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Date         string         `json:"date"`
	ArtistCredit []artistCredit `json:"artist-credit"`
	Genres       []genreItem    `json:"genres"`
	Media        []releaseMedia `json:"media"`
}

type artistCredit struct {
	Name string `json:"name"`
}

type genreItem struct {
	Name string `json:"name"`
}

type releaseMedia struct {
	Position int            `json:"position"`
	Tracks   []releaseTrack `json:"tracks"`
}

type releaseTrack struct {
	Position     int            `json:"position"`
	Number       string         `json:"number"`
	Title        string         `json:"title"`
	ArtistCredit []artistCredit `json:"artist-credit"`
	Recording    struct {
		ID string `json:"id"`
	} `json:"recording"`
}

// EnrichRecording expands a chosen search result into one or more
// release-specific candidates containing album / year / genres / disc / track.
func (c *mbClient) EnrichRecording(
	ctx context.Context,
	rec *Recording,
) ([]EnrichedRecording, error) {
	if rec.ID == "" {
		return nil, fmt.Errorf("recording id is required")
	}

	if len(rec.Releases) == 0 {
		return nil, fmt.Errorf("recording has no releases")
	}

	var out []EnrichedRecording
	seen := make(map[string]struct{})

	for _, rel := range rec.Releases {
		if rel.ID == "" {
			continue
		}

		u := fmt.Sprintf(
			"https://musicbrainz.org/ws/2/release/%s?fmt=json&inc=recordings+artist-credits+genres",
			url.PathEscape(rel.ID),
		)

		var releaseResp releaseLookupResponse
		if err := c.getJSON(ctx, u, &releaseResp); err != nil {
			continue
		}

		for _, media := range releaseResp.Media {
			for _, track := range media.Tracks {
				if track.Recording.ID != rec.ID {
					continue
				}

				trackNo := track.Position
				if trackNo == 0 {
					trackNo = parseIntLoose(track.Number)
				}

				artists := extractNamesFromRecording(rec)
				if len(track.ArtistCredit) > 0 {
					artists = extractArtistNames(track.ArtistCredit)
				}

				item := EnrichedRecording{
					RecordingID:  rec.ID,
					ReleaseID:    releaseResp.ID,
					Title:        rec.Title,
					Artists:      artists,
					Length:       rec.Length,
					Album:        releaseResp.Title,
					AlbumArtists: extractArtistNames(releaseResp.ArtistCredit),
					Year:         parseYear(firstNonEmpty(releaseResp.Date, rel.Date)),
					Genres:       extractGenreNames(releaseResp.Genres),
					TrackNumber:  trackNo,
					DiscNumber:   media.Position,
				}

				key := fmt.Sprintf("%s|%s|%d|%d",
					item.RecordingID,
					item.ReleaseID,
					item.DiscNumber,
					item.TrackNumber,
				)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}

				out = append(out, item)
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no matching release tracks found for recording %s", rec.ID)
	}

	return out, nil
}

func (c *mbClient) getJSON(ctx context.Context, u string, dst any) error {
	<-c.limiter

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("musicbrainz returned %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func extractNamesFromRecording(rec *Recording) []string {
	out := make([]string, 0, len(rec.Artists))
	for _, a := range rec.Artists {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func extractArtistNames(ac []artistCredit) []string {
	out := make([]string, 0, len(ac))
	for _, a := range ac {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func extractGenreNames(gs []genreItem) []string {
	out := make([]string, 0, len(gs))
	seen := make(map[string]struct{})
	for _, g := range gs {
		name := strings.TrimSpace(g.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func parseYear(s string) int {
	if len(s) < 4 {
		return 0
	}
	n, err := strconv.Atoi(s[:4])
	if err != nil {
		return 0
	}
	return n
}

func parseIntLoose(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err == nil {
		return n
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
