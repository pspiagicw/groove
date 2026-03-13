# TODO.md — musicmgr

Comprehensive development checklist for the Go-based music management tool.

---

# Phase 0 — Project Setup

## Repository

* [x] Initialize git repository
* [x] Create Go module (`go mod init musicmgr`)
* [x] Setup basic project structure
* [x] Add `.gitignore`
* [x] Create `README.md`
* [x] Add `TODO.md`

## Project Layout

Create directory structure:

```
musicmgr
├ cmd/
├ internal/
│  ├ config/
│  ├ scanner/
│  ├ metadata/
│  ├ hashing/
│  ├ database/
│  ├ library/
│  ├ playlists/
│  ├ transcode/
│  └ artists/
├ db/
├ main.go
```

---

# Phase 1 — CLI + Config

## CLI Setup

* [x] Add Cobra CLI framework
* [x] Create root command
* [x] Add basic CLI help output
* [x] Implement command registration system

## Config System

* [x] Implement config file loader
* [x] Define config structure
* [x] Support TOML config format
* [ ] Expand `~` paths
* [x] Validate directories
* [x] Handle missing config gracefully

## Config Commands

* [x] `musicmgr init`
* [x] `musicmgr config show`
* [x] `musicmgr config validate`

## Default Config Template

* [x] Generate default config on init

---

# Phase 2 — Database Layer

## Database Setup

* [x] Implement SQLite connection
* [x] Create DB initialization function
* [x] Load schema automatically on startup
* [ ] Implement migrations system (optional)

## Schema Creation

### Artists Table

* [ ] `artists`

### Albums Table

* [ ] `albums`

### Tracks Table

* [ ] `tracks`

### Track Artists (many-to-many)

* [ ] `track_artists`

### Files Table

* [ ] `files`

### Import Queue

* [ ] `import_queue`

## Database Utilities

* [ ] Insert functions
* [ ] Query helpers
* [ ] Transaction helpers
* [ ] Indexes for performance

---

# Phase 3 — File Scanner

## File Discovery

* [ ] Recursive directory scanning
* [ ] Filter supported extensions

Supported formats:

* [ ] mp3
* [ ] flac
* [ ] opus
* [ ] ogg
* [ ] m4a
* [ ] wav

## Scanner Tasks

* [ ] Detect new files
* [ ] Ignore already imported files
* [ ] Ignore unsupported formats
* [ ] Handle broken files

## Hash System

* [ ] Implement file hashing
* [ ] Use SHA256
* [ ] Detect duplicates using hash
* [ ] Store hash in database

## Queue Insert

* [ ] Insert scanned files into `import_queue`

## CLI

* [ ] `musicmgr scan`
* [ ] `--force`
* [ ] `--ext`

---

# Phase 4 — Metadata Extraction

## Metadata Reader

* [ ] Integrate tag reader library
* [ ] Extract metadata fields

Fields:

* [ ] title
* [ ] artist
* [ ] album
* [ ] track number
* [ ] disc number
* [ ] year
* [ ] genre
* [ ] album art

## Metadata Normalization

* [ ] Trim whitespace
* [ ] Normalize casing
* [ ] Handle missing metadata
* [ ] Split multi-artist tags

---

# Phase 5 — Artist Parsing

## Artist Splitter

Support parsing formats:

* [ ] `Artist1 & Artist2`
* [ ] `Artist1, Artist2`
* [ ] `Artist1 feat Artist2`
* [ ] `Artist1 ft Artist2`
* [ ] `Artist1 featuring Artist2`

## Artist Normalization

* [ ] Remove duplicates
* [ ] Standardize formatting

## Database Storage

* [ ] Store artists individually
* [ ] Link via `track_artists` table

---

# Phase 6 — Import Queue Management

## Queue Listing

* [ ] Show pending files
* [ ] Show duplicate entries
* [ ] Show metadata status

## Queue Commands

### List

* [ ] `musicmgr queue list`

### Show

* [ ] `musicmgr queue show <id>`

### Edit

* [ ] `musicmgr queue edit <id>`

Editable fields:

* [ ] artist
* [ ] album
* [ ] title
* [ ] genre
* [ ] year
* [ ] track number

### Remove

* [ ] `musicmgr queue remove <id>`

---

# Phase 7 — Import System

## Import Pipeline

Steps:

* [ ] Load queue entry
* [ ] Validate metadata
* [ ] Generate destination path
* [ ] Create directory structure
* [ ] Rename file
* [ ] Move file to library
* [ ] Insert records into DB
* [ ] Mark queue item as imported

## Naming Scheme

Support format strings:

Examples:

```
{artist}/{album}/{track:02} - {title}.mp3
```

Fields supported:

* [ ] artist
* [ ] album
* [ ] track
* [ ] title
* [ ] year
* [ ] genre

## Import Command

* [ ] `musicmgr import`

Flags:

* [ ] `--id`
* [ ] `--dry-run`
* [ ] `--limit`

---

# Phase 8 — Duplicate Detection

## Hash Matching

* [ ] Compare file hashes
* [ ] Skip duplicates

## Duplicate Handling

* [ ] Mark duplicate in queue
* [ ] Allow override import

Future:

* [ ] Audio fingerprint detection

---

# Phase 9 — Playlist Generator

## Playlist Types

### Artist Playlists

* [ ] Generate playlists per artist
* [ ] Save as `.m3u`

Example:

```
playlists/artists/radiohead.m3u
```

### Genre Playlists

* [ ] Generate playlists per genre

Example:

```
playlists/genres/jazz.m3u
```

## Playlist Command

* [ ] `musicmgr playlists generate`

Flags:

* [ ] `--artists`
* [ ] `--genres`

---

# Phase 10 — Transcoding System

## Transcoding Engine

Use external tools:

* [ ] ffmpeg
* [ ] lame
* [ ] opusenc

## Codec Support

Input:

* [ ] opus
* [ ] flac
* [ ] wav

Output:

* [ ] mp3 (initial)
* [ ] configurable later

## Transcode Command

* [ ] `musicmgr transcode`

Flags:

* [ ] `--codec`
* [ ] `--bitrate`
* [ ] `--dry-run`

---

# Phase 11 — Playlist Metadata Improvements

* [ ] Group artists properly
* [ ] Handle multi-genre tracks
* [ ] Generate smart playlists
* [ ] Deduplicate entries

---

# Phase 12 — Library Statistics

## Stats Command

* [ ] `musicmgr stats`

Display:

* [ ] total artists
* [ ] total albums
* [ ] total tracks
* [ ] total genres
* [ ] total size
* [ ] codec distribution

---

# Phase 13 — Error Handling

* [ ] Corrupt files
* [ ] Missing metadata
* [ ] filesystem permission errors
* [ ] interrupted imports
* [ ] DB failures

---

# Phase 14 — Logging

* [ ] Add structured logging
* [ ] Debug logging
* [ ] Error logs
* [ ] Import logs

---

# Phase 15 — Performance Improvements

* [ ] Parallel scanning
* [ ] Batch DB inserts
* [ ] Cache hash results
* [ ] Reduce disk IO

---

# Phase 16 — Future Features

## Audio Fingerprinting

* [ ] Chromaprint integration

## Online Metadata

* [ ] MusicBrainz lookup
* [ ] Cover art download

## Advanced Playlists

* [ ] smart playlists
* [ ] recently added
* [ ] top artists

## Audio Features

* [ ] ReplayGain
* [ ] loudness normalization

---

# Phase 17 — Testing

## Unit Tests

* [ ] scanner tests
* [ ] metadata tests
* [ ] hashing tests
* [ ] database tests

## Integration Tests

* [ ] full import pipeline
* [ ] duplicate detection
* [ ] playlist generation

---

# Phase 18 — Documentation

* [ ] CLI usage guide
* [ ] config file reference
* [ ] library structure docs
* [ ] developer guide

---

# Phase 19 — Packaging

* [ ] Build static binaries
* [ ] Linux release
* [ ] macOS release
* [ ] Windows release (optional)

---

# Long-Term Vision

Potential future improvements:

* [ ] web UI
* [ ] TUI interface
* [ ] MPD integration
* [ ] mobile sync
* [ ] streaming server
* [ ] waveform analysis
* [ ] audio similarity clustering

