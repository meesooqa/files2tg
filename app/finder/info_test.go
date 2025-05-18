package finder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestVideoInfoProvider struct{}

func NewTestVideoInfoProvider() *TestVideoInfoProvider {
	return &TestVideoInfoProvider{}
}

func (o *TestVideoInfoProvider) GetVideoInfo(path string) (*VideoInfo, error) {
	return nil, nil
}

// createFakeFFProbe installs a temporary "ffprobe" script at the front of PATH
// which outputs the provided JSON and ignores its arguments.
func createFakeFFProbe(t *testing.T, output string) {
	t.Helper()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "ffprobe")
	// Write a shell script that prints the desired JSON
	script := []byte("#!/bin/sh\ncat <<EOF\n" + output + "\nEOF\n")
	require.NoError(t, os.WriteFile(scriptPath, script, 0755))
	// Prepend tmpDir to PATH so our fake ffprobe is used first
	existing := os.Getenv("PATH")
	require.NoError(t, os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+existing))
}

func TestNewVideoInfoFromFilepath_Success(t *testing.T) {
	// Given valid ffprobe JSON for a video stream
	jsonOutput := `{"streams":[{"codec_type":"video","duration":"4.983244","width":1920,"height":1080}]}`
	createFakeFFProbe(t, jsonOutput)

	vip := NewVideoInfoProvider()
	vid, err := vip.GetVideoInfo("dummy.mp4")

	require.NoError(t, err)
	require.NotNil(t, vid)
	require.Equal(t, "video", vid.CodecType)
	require.Equal(t, "4.983244", vid.DurationRaw)
	require.Equal(t, 1920, vid.Width)
	require.Equal(t, 1080, vid.Height)
	// Duration should be truncated to integer part
	require.Equal(t, 4, vid.Duration)
}

func TestNewVideoInfoFromFilepath_NoVideoStream(t *testing.T) {
	jsonOutput := `{"streams":[{"codec_type":"audio","duration":"3.14","width":0,"height":0}]}`
	createFakeFFProbe(t, jsonOutput)

	vip := NewVideoInfoProvider()
	vid, err := vip.GetVideoInfo("dummy.mp4")

	require.Error(t, err)
	require.Contains(t, err.Error(), "video stream not found")
	require.Nil(t, vid)
}

func TestNewVideoInfoFromFilepath_InvalidDuration(t *testing.T) {
	jsonOutput := `{"streams":[{"codec_type":"video","duration":"notafloat","width":1280,"height":720}]}`
	createFakeFFProbe(t, jsonOutput)

	vip := NewVideoInfoProvider()
	vid, err := vip.GetVideoInfo("dummy.mp4")

	require.Error(t, err)
	require.Contains(t, err.Error(), "ParseFloat")
	require.Nil(t, vid)
}
