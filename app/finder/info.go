package finder

import (
	"encoding/json"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

type VIProvider interface {
	GetVideoInfo(path string) (*VideoInfo, error)
}

type FFProbe struct {
	Streams []VideoInfo `json:"streams"`
}

// VideoInfo involves video file info
type VideoInfo struct {
	CodecType   string `json:"codec_type"`
	DurationRaw string `json:"duration"`

	Width  int `json:"width"`
	Height int `json:"height"`
	// Duration of the recording in seconds
	Duration int
}

type VideoInfoProvider struct{}

func NewVideoInfoProvider() *VideoInfoProvider {
	return &VideoInfoProvider{}
}

func (o *VideoInfoProvider) GetVideoInfo(path string) (*VideoInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		"-of", "json", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info FFProbe
	if err = json.Unmarshal(out, &info); err != nil {
		return nil, err
	}

	var vid *VideoInfo
	for i := range info.Streams {
		if info.Streams[i].CodecType == "video" {
			vid = &info.Streams[i]
			break
		}
	}
	if vid == nil {
		return nil, errors.New("video stream not found")
	}

	durationFloat, err := strconv.ParseFloat(vid.DurationRaw, 64)
	if err != nil {
		return nil, err
	}
	vid.Duration = int(durationFloat)

	return vid, nil
}
