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
	Info    *VideoInfo
}

type Provider struct {
	VideoInfoProvider VIProvider
}

func NewProvider(VideoInfoProvider VIProvider) *Provider {
	return &Provider{
		VideoInfoProvider: VideoInfoProvider,
	}
}

// GetListFilesSorted returns a list of files in a directory
// sorted by modification time
func (o *Provider) GetListFilesSorted(root, dir string) ([]File, error) {
	return o.listFilesSorted(os.DirFS(root), root, dir)
}

// ListFilesSorted returns a list of files in a directory
// sorted by modification time
func (o *Provider) listFilesSorted(fsys fs.FS, root, dir string) ([]File, error) {
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
		path := filepath.Join(root, dir, entry.Name())
		videoInfo, err := o.VideoInfoProvider.GetVideoInfo(path)
		if err != nil {
			continue
			// return nil, fmt.Errorf("failed to get videoInfo for %s: %w", path, err)
		}
		files = append(files, File{
			Name:    entry.Name(),
			ModTime: info.ModTime(),
			Path:    path,
			Info:    videoInfo,
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})
	return files, nil
}
