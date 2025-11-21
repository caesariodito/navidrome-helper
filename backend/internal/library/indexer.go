package library

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"navidrome-helper/internal/config"
	"navidrome-helper/internal/store"
	"navidrome-helper/internal/util"
)

// Indexer scans NAVIDROME_MUSIC_PATH and writes results to SQLite.
type Indexer struct {
	cfg   config.Config
	store *store.Store
}

func NewIndexer(cfg config.Config, store *store.Store) *Indexer {
	return &Indexer{cfg: cfg, store: store}
}

// Refresh scans the filesystem and replaces the library index.
func (i *Indexer) Refresh(ctx context.Context) ([]store.LibraryEntry, error) {
	var entries []store.LibraryEntry
	artistDirs, err := os.ReadDir(i.cfg.NavidromePath)
	if err != nil {
		return nil, fmt.Errorf("read navidrome path: %w", err)
	}
	now := time.Now().UTC()
audioLoop:
	for _, artist := range artistDirs {
		if !artist.IsDir() || strings.HasPrefix(artist.Name(), ".") {
			continue
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		artistPath := filepath.Join(i.cfg.NavidromePath, artist.Name())
		albumDirs, err := os.ReadDir(artistPath)
		if err != nil {
			continue audioLoop
		}
		for _, album := range albumDirs {
			if !album.IsDir() || strings.HasPrefix(album.Name(), ".") {
				continue
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			albumPath := filepath.Join(artistPath, album.Name())
			count, err := countAudioFiles(albumPath)
			if err != nil {
				continue
			}
			artistNorm := util.NormalizeName(artist.Name())
			albumNorm := util.NormalizeName(album.Name())
			entry := store.LibraryEntry{
				Artist:     artist.Name(),
				Album:      album.Name(),
				Path:       albumPath,
				TrackCount: count,
				UpdatedAt:  now,
				artistNorm: artistNorm,
				albumNorm:  albumNorm,
			}
			entries = append(entries, entry)
		}
	}
	if err := i.store.ReplaceLibraryIndex(entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func countAudioFiles(root string) (int, error) {
	audioExts := map[string]struct{}{
		".mp3": {}, ".flac": {}, ".ogg": {}, ".wav": {}, ".alac": {}, ".aac": {}, ".m4a": {},
	}
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := audioExts[ext]; ok {
			count++
		}
		return nil
	})
	return count, err
}
