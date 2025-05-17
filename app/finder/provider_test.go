package finder

import (
	"fmt"
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

func TestListFilesSortedAndChunked(t *testing.T) {
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

		p := NewProvider()
		chunks, err := p.listFilesSortedAndChunked(fsys, "", ".", 10)
		require.NoError(t, err)
		require.Len(t, chunks, 1)
		require.Len(t, chunks[0], 3)

		expectedOrder := []string{"file2.txt", "file1.txt", "file3.txt"}
		for i, name := range expectedOrder {
			assert.Equal(t, name, chunks[0][i].Name)
		}
	})

	t.Run("chunking with more than chunkSize virtual files", func(t *testing.T) {
		fsMap := make(fstest.MapFS)
		for i := 1; i <= 10; i++ {
			fsMap[fmt.Sprintf("file%d.txt", i)] = &fstest.MapFile{
				Data:    []byte("content"),
				ModTime: time.Date(2023, 1, int(time.Month(i)), 0, 0, 0, 0, time.UTC),
			}
		}

		p := NewProvider()
		chunks, err := p.listFilesSortedAndChunked(fsMap, "", ".", 7)
		require.NoError(t, err)
		assert.Len(t, chunks, 2)
		assert.Len(t, chunks[0], 7)
		assert.Len(t, chunks[1], 3)
	})

	t.Run("invalid directory", func(t *testing.T) {
		fsys := fstest.MapFS{}
		p := NewProvider()
		chunks, err := p.listFilesSortedAndChunked(fsys, "", "nonexistent", 5)
		assert.Error(t, err)
		assert.Nil(t, chunks)
	})
}
