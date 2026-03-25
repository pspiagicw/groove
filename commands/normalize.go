package commands

import "strings"

func normalizeTitle(title string) string {
	// TODO: Implement (feat. ) and (PROD) normalization.
	return title
}

func normalizeArtist(artist []string) []string {
	newArtists := []string{}
	for _, a := range artist {
		// TODO: Implement normalization
		for newA := range strings.SplitSeq(a, ";") {
			newArtists = append(newArtists, strings.TrimSpace(newA))
		}
	}
	return newArtists
}

func denormalizeArtist(artist []string) string {
	return strings.Join(artist, ";")
}
func normalizeAlbum(name string) string {
	// TODO: Implement (Deluxe and other removal) album normalization.
	return name
}

func normalizeAlbumArtist(name string) string {
	// TODO: Implement album artist normalization.
	genres := strings.Split(name, ",")
	for i, genre := range genres {
		genres[i] = strings.TrimSpace(genre)
	}
	return strings.Join(genres, " · ")
	return name
}

func normalizeGenre(name string) string {
	// TODO: Implement genre normalization, including genre folding.
	genres := strings.Split(name, "/")
	for i, genre := range genres {
		genres[i] = strings.TrimSpace(genre)
	}
	return strings.Join(genres, " · ")
}

func denormalizeGenre(name string) string {
	return strings.Join(strings.Split(name, " · "), "/")
}

func normalizeYear(year int) int {
	// TODO: Implement year normalization
	return year
}

func normalizeTrackNumber(track int) int {
	// TODO: Implement track normalization (like adding zeroes and stuff).
	return track
}

func normalizeDiscNumber(disc int) int {
	// TODO: Disc Normalization
	return disc
}
