package finder

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualFSWithDir(t *testing.T) {
	fsys := fstest.MapFS{
		"file.txt": {
			Data:    []byte("content"),
			ModTime: time.Now(),
			Mode:    0644,
		},
		"dir1/": {
			Data:    nil,
			ModTime: time.Now(),
			Mode:    os.ModeDir, // IsDir
		},
	}

	info, err := fs.Stat(fsys, "dir1")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestListFilesSorted(t *testing.T) {
	t.Run("sorted and filtered virtual files", func(t *testing.T) {
		modTime := func(year int) time.Time {
			return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		fsys := fstest.MapFS{
			"file1.txt":        {Data: []byte("content"), ModTime: modTime(2023)},
			"file2.txt":        {Data: []byte("content"), ModTime: modTime(2022)},
			"file3.txt":        {Data: []byte("content"), ModTime: modTime(2024)},
			"dir1/":            {Data: nil, Mode: os.ModeDir, ModTime: modTime(2021)},
			"dir1/subfile.txt": {Data: []byte("content"), ModTime: modTime(2020)},
		}

		testVIP := NewTestVideoInfoProvider()
		p := NewProvider(testVIP)
		files, err := p.listFilesSorted(fsys, "", ".")
		require.NoError(t, err)
		require.Len(t, files, 3)

		expectedOrder := []string{"file2.txt", "file1.txt", "file3.txt"}
		for i, name := range expectedOrder {
			assert.Equal(t, name, files[i].Name)
		}
	})

	t.Run("invalid directory", func(t *testing.T) {
		fsys := fstest.MapFS{}
		testVIP := NewTestVideoInfoProvider()
		p := NewProvider(testVIP)
		files, err := p.listFilesSorted(fsys, "", "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, files)
	})
}
