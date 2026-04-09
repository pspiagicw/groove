package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/pspiagicw/groove/database"
	"github.com/pspiagicw/groove/musicbrainz"
	"github.com/pspiagicw/groove/prettylog"
)

const MISSING_PLACEHOLDER = "<missing>"

type ImportSession struct {
	database.ScannedItem

	NormalizedTitle       string
	NormalizedArtists     []string
	NormalizedAlbum       string
	NormalizedAlbumArtist []string
	NormalizedYear        int
	NormalizedGenre       string
	NormalizedTrackNumber int
	NormalizedDiscNumber  int

	CurrentTitle       string
	CurrentArtists     []string
	CurrentAlbum       string
	CurrentAlbumArtist []string
	CurrentYear        int
	CurrentGenre       string
	CurrentTrackNumber int
	CurrentDiscNumber  int

	MusicBrainzCandidates        []musicbrainz.Recording
	SelectedMusicBrainzCandidate musicbrainz.Recording
	Skipped                      bool
}

func confirmItemDetails(info *database.ScannedItem) ImportSession {
	item := newSession(info)

	item = normalize(item)

	item = startSession(item)

	return item
}

func startSession(i ImportSession) ImportSession {
	return showReviewScreen(i)

}
func showReviewScreen(i ImportSession) ImportSession {

	displayCurrentDetails(i)
	displayChangedDetails(i)

	skipIndex := 0

	if fieldsMissing(i) {
		prettylog.Infof("Fields are missing: correct them to include into database.")
		skipIndex = 1
	}

	result := promptUser([]string{
		"Accept and Import",
		"Edit Manually",
		"Lookup Online",
		"Skip",
	}, skipIndex)

	// These should match above order.
	switch result {
	case 1:
		return importSuccess(i)
	case 2:
		return editManually(i)
	case 3:
		return lookupPrecheck(i)
	case 4:
		return skipItem(i)
	default:
		return showReviewScreen(i)
	}
}

func skipItem(i ImportSession) ImportSession {
	i.Skipped = true
	return i
}

func importSuccess(i ImportSession) ImportSession {
	return i
}
func lookupPrecheck(i ImportSession) ImportSession {
	if i.CurrentTitle == MISSING_PLACEHOLDER || len(i.CurrentArtists) == 0 {
		return lookupBlocked(i)
	}
	// TODO:
	return musicBrainzSearch(i)
}

func musicBrainzSearch(i ImportSession) ImportSession {
	result, err := musicbrainz.Query(i.NormalizedTitle, strings.Join(i.NormalizedArtists, ","))

	if err != nil {
		prettylog.Errorf("Failed to query musicbrainz: %v!", err)
	}

	if result == nil {
		return musicBrainzNoResults(i)
	}

	return showMusicBrainzResults(i, result)
}
func artistsToString(artists []struct{ Name string }) string {
	names := []string{}
	for _, artist := range artists {
		names = append(names, artist.Name)
	}
	return strings.Join(names, " · ")
}

// func releaseToString(releases []struct{Title string Date string}) string {
// 	return string(len(releases))
// }

func formatRelease(releases []musicbrainz.Release) string {
	if len(releases) < 1 {
		prettylog.Errorf("Not 1 release, %d", len(releases))
		return "0"
	}

	item := strconv.Itoa(len(releases))

	// TODO: Add different style later on.
	return item
}

func formatRecording(r musicbrainz.Recording) string {
	return r.Title + " - " + artistsToString([]struct{ Name string }(r.Artists)) + " - " + formatRelease(r.Releases)
}
func chooseRecording(results []musicbrainz.Recording) musicbrainz.Recording {
	choices := []huh.Option[int]{}
	var choice int
	for i, result := range results {
		choices = append(choices, huh.NewOption(formatRecording(result), i))

	}

	err := huh.NewSelect[int]().Title("Choose recording").Options(choices...).Value(&choice).Run()

	if err != nil {
		prettylog.Fatalf("Failed to run form: %v!", err)
	}

	return results[choice]
}

func showMusicBrainzResults(i ImportSession, result []musicbrainz.Recording) ImportSession {
	recording := chooseRecording(result)

	return useMusicBrainzResult(&recording, i)
}
func useMusicBrainzResult(result *musicbrainz.Recording, i ImportSession) ImportSession {
	ctx := context.Background()

	client := musicbrainz.NewMBClient("groove/0.1 {https://github.com/pspiagicw/groove}")

	candidates, err := client.EnrichRecording(ctx, result)

	if err != nil {
		prettylog.Errorf("Failed to get more details for recording: %v", err)
	}

	r := chooseCandidate(candidates, i)

	choice := promptUser([]string{
		"Apply All",
		"Apply Selectively",
		"Edit Manually",
		"Back",
	}, 0)

	switch choice {
	case 1:
		return applyAll(i, r)
	case 2:
		return applySelectively(i, r)
	case 3:
		return editManually(i)
	case 4:
		return showReviewScreen(i)
	}

	return showReviewScreen(i)
}

func listToString(artists []string) string {
	mods := []string{}
	for _, a := range artists {
		mods = append(mods, fmt.Sprintf("'%s'", a))
	}

	return strings.Join(mods, " · ")
}
func applyAll(i ImportSession, r *musicbrainz.EnrichedRecording) ImportSession {
	i.NormalizedTitle = r.Title
	i.NormalizedArtists = r.Artists
	i.NormalizedAlbum = r.Album
	i.NormalizedAlbumArtist = r.AlbumArtists
	i.NormalizedYear = r.Year
	i.NormalizedTrackNumber = r.TrackNumber
	i.NormalizedDiscNumber = r.DiscNumber

	return showReviewScreen(i)
}
func applySelectively(i ImportSession, r *musicbrainz.EnrichedRecording) ImportSession {
	i.NormalizedTitle = r.Title
	i.NormalizedArtists = r.Artists
	i.NormalizedAlbum = r.Album
	i.NormalizedAlbumArtist = r.AlbumArtists
	i.NormalizedDiscNumber = r.DiscNumber
	i.NormalizedTrackNumber = r.TrackNumber
	i.NormalizedYear = r.Year
	return editManually(i)
}

func chooseCandidate(candidates []musicbrainz.EnrichedRecording, i ImportSession) *musicbrainz.EnrichedRecording {

	// inputRange := len(candidates)

	// for _, candidate := range candidates {
	// 	displayEnrichedRecording(candidate)
	// }
	choices := []string{}
	for _, c := range candidates {
		str := strings.Join(
			[]string{
				c.Title,
				listToString(c.Artists),
				c.Album,
				listToString(c.AlbumArtists),
				listToString(c.Genres),
				string(c.Year),
				string(c.TrackNumber),
				string(c.DiscNumber),
			}, " - ")
		choices = append(choices, string(str))
	}

	choice := promptUser(choices, 0)
	// choice := promptUserWithNumber([]string{
	// 	"Back",
	// }, inputRange)
	//
	// if choice > 0 && choice <= inputRange {
	// 	return &candidates[choice-1]
	// } else if choice == inputRange+1 {
	// 	showReviewScreen(i)
	// } else {
	// 	return chooseCandidate(candidates, i)
	// }
	displayEnrichedRecording(candidates[choice-1])
	return &candidates[choice-1]

}
func editManually(i ImportSession) ImportSession {
	artists := denormalizeArtist(i.NormalizedArtists)
	genre := denormalizeGenre(i.NormalizedGenre)
	albumArtists := denormalizeAlbumArtists(i.NormalizedAlbumArtist)
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Value(&i.NormalizedTitle),
			huh.NewInput().Title("Artists").Value(&artists),
		),
		huh.NewGroup(
			huh.NewInput().Title("Album").Value(&i.NormalizedAlbum),
			huh.NewInput().Title("Album Artist").Value(&albumArtists),
			huh.NewInput().Title("Genre").Value(&genre),
		),
	).Run()

	if err != nil {
		prettylog.Errorf("Failed to execute form: %v", err)
	}

	i.NormalizedArtists = normalizeArtist([]string{artists})
	i.NormalizedGenre = normalizeGenre(genre)
	i.NormalizedAlbumArtist = normalizeAlbumArtists([]string{albumArtists})

	return showReviewScreen(i)
}

func musicBrainzNoResults(i ImportSession) ImportSession {
	result := promptUser([]string{
		"Edit Manually",
		"Retry Search",
		"Back",
	}, 0)

	switch result {
	case 1:
		return editManually(i)
	case 2:
		return musicBrainzSearch(i)
	case 3:
		return showReviewScreen(i)
	default:
		return musicBrainzNoResults(i)
	}
}
func lookupBlocked(i ImportSession) ImportSession {
	prettylog.Infof("Artist and Title are required fields for online lookup, please add them manually for online lookup.")

	result := promptUser([]string{
		"Edit Manually",
		"Back",
	}, 0)

	switch result {
	case 1:
		return editManually(i)
	case 2:
		return showReviewScreen(i)
	default:
		return lookupBlocked(i)
	}
}

func fieldsMissing(i ImportSession) bool {
	missing := false
	if i.NormalizedTitle == MISSING_PLACEHOLDER {
		missing = true
	}
	if len(i.NormalizedArtists) == 0 {
		missing = true
	}
	if i.NormalizedAlbum == MISSING_PLACEHOLDER {
		missing = true
	}

	if len(i.NormalizedAlbumArtist) == 0 {
		missing = true
	}
	if i.NormalizedGenre == MISSING_PLACEHOLDER {
		missing = true
	}

	return missing
}

func newSession(info *database.ScannedItem) ImportSession {
	item := new(ImportSession)
	item.ScannedItem = *info

	item.CurrentTitle = checkIfMissing(info.DetectedTitle)
	item.CurrentArtists = []string{info.DetectedArtist}
	if info.DetectedArtist == "" {
		item.CurrentArtists = []string{}
	}
	item.CurrentAlbum = checkIfMissing(info.DetectedAlbum)

	item.CurrentAlbumArtist = []string{info.DetectedAlbumArtist}
	if info.DetectedAlbumArtist == "" {
		item.CurrentAlbumArtist = []string{}
	}
	item.CurrentGenre = checkIfMissing(info.DetectedGenre)

	item.CurrentYear = info.DetectedYear
	item.CurrentTrackNumber = info.DetectedTrackNumber
	item.CurrentDiscNumber = info.DetectedDisc

	return *item
}
func checkIfMissing(item string) string {
	if item == "" {
		return MISSING_PLACEHOLDER
	}

	return item
}
func normalize(i ImportSession) ImportSession {

	i.NormalizedTitle = normalizeTitle(i.CurrentTitle)
	i.NormalizedArtists = normalizeArtist(i.CurrentArtists)
	i.NormalizedAlbum = normalizeAlbum(i.CurrentAlbum)
	i.NormalizedAlbumArtist = normalizeAlbumArtists(i.CurrentAlbumArtist)
	i.NormalizedGenre = normalizeGenre(i.CurrentGenre)
	i.NormalizedYear = normalizeYear(i.CurrentYear)
	i.NormalizedTrackNumber = normalizeTrackNumber(i.CurrentTrackNumber)
	i.NormalizedDiscNumber = normalizeDiscNumber(i.CurrentDiscNumber)

	return i
}

// Display the current song with current details.
func displayCurrentDetails(info ImportSession) {
	prettylog.PrintBlock(
		os.Stdout,
		"Item to Import",
		prettylog.KV("Filepath", info.Path),
		prettylog.KV("Title", info.CurrentTitle),
		prettylog.KV("Artists", info.CurrentArtists),
		prettylog.KV("Album Artist", info.CurrentAlbumArtist),
		prettylog.KV("Album", info.CurrentAlbum),
		prettylog.KV("Year", info.CurrentYear),
		prettylog.KV("Genre", info.CurrentGenre),
		prettylog.KV("TrackNumber", info.CurrentTrackNumber),
		prettylog.KV("DiscNumber", info.CurrentDiscNumber),
		prettylog.KV("Filetype", info.DetectedFileType),
	)
}

func displayEnrichedRecording(c musicbrainz.EnrichedRecording) {
	prettylog.PrintBlock(
		os.Stdout,
		"Item",
		prettylog.KV("Title", c.Title),
		prettylog.KV("Artists", c.Artists),
		prettylog.KV("Album", c.Album),
		prettylog.KV("Album Artist", c.AlbumArtists),
		prettylog.KV("Year", c.Year),
		prettylog.KV("Genre", c.Genres),
		prettylog.KV("TrackNumber", c.TrackNumber),
		prettylog.KV("DiscNumber", c.DiscNumber),
	)
}

func compareArtists(oldArtists []string, newArtists []string) bool {
	if len(oldArtists) != len(newArtists) {
		return false
	}
	for i, old := range oldArtists {
		if old != newArtists[i] {
			return false
		}
	}
	return true

}

// Display the changes (difference between current and normalized.)
func displayChangedDetails(info ImportSession) {
	changedFields := []string{}

	if info.NormalizedTitle != info.CurrentTitle {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Title",
				info.CurrentTitle,
				info.NormalizedTitle,
			))
	}

	if !compareArtists(info.CurrentArtists, info.NormalizedArtists) {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Artists",
				info.CurrentArtists,
				info.NormalizedArtists,
			))
	}

	if info.NormalizedAlbum != info.CurrentAlbum {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Album",
				info.CurrentAlbum,
				info.NormalizedAlbum,
			))
	}
	if !compareArtists(info.CurrentAlbumArtist, info.NormalizedAlbumArtist) {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Album Artists",
				info.CurrentAlbumArtist,
				info.NormalizedAlbumArtist,
			))
	}
	// if info.NormalizedAlbumArtist != info.CurrentAlbumArtist {
	// 	changedFields = append(
	// 		changedFields,
	// 		prettylog.KVWithDiff(
	// 			"AlbumArtist",
	// 			info.CurrentAlbumArtist,
	// 			info.NormalizedAlbumArtist,
	// 		))
	// }
	if info.NormalizedGenre != info.CurrentGenre {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Genre",
				info.CurrentGenre,
				info.NormalizedGenre,
			))
	}
	if info.NormalizedYear != info.CurrentYear {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Year",
				info.CurrentYear,
				info.NormalizedYear,
			))
	}
	if info.NormalizedDiscNumber != info.CurrentDiscNumber {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Disc",
				info.CurrentDiscNumber,
				info.NormalizedDiscNumber,
			))
	}
	if info.NormalizedTrackNumber != info.CurrentTrackNumber {
		changedFields = append(
			changedFields,
			prettylog.KVWithDiff(
				"Track",
				info.CurrentTrackNumber,
				info.NormalizedTrackNumber,
			))
	}

	if len(changedFields) == 0 {
		return
	}

	prettylog.PrintBlock(
		os.Stdout,
		"Fields with change",
		changedFields...,
	)
}

func confirm(message string) bool {
	var result bool
	err := huh.NewConfirm().Title(message).Affirmative("yes!").Negative("NO!!").Value(&result).Run()
	if err != nil {
		prettylog.Fatalf("Error running prompt: %v!", err)
	}
	return result
}

// Prompt user with choices, return value should be >= 1
func promptUser(choices []string, skipIndex int) int {
	options := []huh.Option[int]{}
	var choiceIndex int
	for i, choice := range choices {
		if i == skipIndex-1 {
			continue
		}
		options = append(options, huh.NewOption(choice, i+1))
	}

	err := huh.NewSelect[int]().Title("Choose a option").Options(options...).Value(&choiceIndex).Run()

	if err != nil {
		prettylog.Fatalf("Failed to run form: %v", err)
	}

	return choiceIndex
}

// START
//   ↓
// LOAD_QUEUE_ITEM
//   ↓
// NORMALIZE_METADATA
//   ↓
// SHOW_REVIEW_SCREEN
//   ├── ACCEPT_CURRENT → SHOW_FINAL_CONFIRMATION
//   ├── EDIT_MANUALLY → EDIT_MENU
//   ├── LOOKUP_ONLINE → LOOKUP_PRECHECK
//   ├── SKIP_ITEM → SKIP_CONFIRMATION
//   └── MARK_DUPLICATE → DUPLICATE_CONFIRMATION
//
// LOOKUP_PRECHECK
//   ├── ENOUGH_DATA → MUSICBRAINZ_SEARCH
//   └── NOT_ENOUGH_DATA → LOOKUP_BLOCKED_PROMPT
//
// MUSICBRAINZ_SEARCH
//   ├── NO_RESULTS → LOOKUP_NO_RESULTS_PROMPT
//   └── RESULTS_FOUND → SHOW_MB_RESULTS
//
// SHOW_MB_RESULTS
//   ├── SELECT_RESULT → SHOW_MB_APPLY_PREVIEW
//   ├── MANUAL_EDIT → EDIT_MENU
//   ├── RETRY_SEARCH → MUSICBRAINZ_SEARCH
//   └── BACK → SHOW_REVIEW_SCREEN
//
// SHOW_MB_APPLY_PREVIEW
//   ├── APPLY_ALL → UPDATE_CURRENT_METADATA
//   ├── APPLY_SELECTIVE → SELECTIVE_APPLY_MENU
//   └── CANCEL → SHOW_MB_RESULTS
//
// EDIT_MENU
//   ├── EDIT_TITLE
//   ├── EDIT_ARTISTS
//   ├── EDIT_ALBUM
//   ├── EDIT_ALBUM_ARTISTS
//   ├── EDIT_OPTIONAL_FIELDS
//   ├── DONE_EDITING → SHOW_REVIEW_SCREEN
//   └── CANCEL → SHOW_REVIEW_SCREEN
//
// SHOW_FINAL_CONFIRMATION
//   ├── CONFIRM_IMPORT → EXECUTE_IMPORT
//   ├── GO_BACK → SHOW_REVIEW_SCREEN
//   └── CANCEL → SHOW_REVIEW_SCREEN
//
// EXECUTE_IMPORT
//   ├── SUCCESS → IMPORT_SUCCESS
//   └── FAILURE → IMPORT_ERROR
//
// IMPORT_SUCCESS → NEXT_ITEM / END
// IMPORT_ERROR → ERROR_MENU
