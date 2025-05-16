package finder

import (
	"fmt"
	"io/fs"
	"sort"
	"time"
)

// File involves file info
type File struct {
	Name    string
	ModTime time.Time
}

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

// ListFilesSortedAndChunked returns a list of files in a directory
// sorted by modification time and split into chunks
func (o *Provider) ListFilesSortedAndChunked(fsys fs.FS, dir string, chunkSize int) ([][]File, error) {
	// dir := "."
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []File
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get info for %s: %w", entry.Name(), err)
		}
		files = append(files, File{
			Name:    entry.Name(),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	var chunks [][]File
	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize
		if end > len(files) {
			end = len(files)
		}
		chunks = append(chunks, files[i:end])
	}

	return chunks, nil
}
