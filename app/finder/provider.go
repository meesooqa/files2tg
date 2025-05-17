package finder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// File involves file info
type File struct {
	Path    string
	Name    string
	ModTime time.Time
}

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

// GetListFilesSortedAndChunked returns a list of files in a directory
// sorted by modification time and split into chunks
func (o *Provider) GetListFilesSortedAndChunked(root, dir string, chunkSize int) ([][]File, error) {
	return o.listFilesSortedAndChunked(os.DirFS(root), root, dir, chunkSize)
}

// ListFilesSortedAndChunked returns a list of files in a directory
// sorted by modification time and split into chunks
func (o *Provider) listFilesSortedAndChunked(fsys fs.FS, root, dir string, chunkSize int) ([][]File, error) {
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
			Path:    filepath.Join(root, dir, entry.Name()),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})
	return o.chunk(files, chunkSize), nil
}

func (o *Provider) chunk(files []File, chunkSize int) [][]File {
	var chunks [][]File
	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize
		if end > len(files) {
			end = len(files)
		}
		chunks = append(chunks, files[i:end])
	}
	return chunks
}
