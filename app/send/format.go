package send

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/meesooqa/files2tg/app/finder"
)

type TelegramFormatter struct{}

// Format generates HTML message from provided finder.File
func (o *TelegramFormatter) Format(file finder.File) string {
	// file.Name == "Секретная инфа [3873955765].mp4"
	ext := filepath.Ext(file.Name)
	base := strings.TrimSuffix(file.Name, ext)
	re := regexp.MustCompile(`^(?P<name>.*)\s*\[(?P<id>\d+)\]$`)
	var id, title string
	matches := re.FindStringSubmatch(base)
	if matches != nil {
		result := map[string]string{}
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = matches[i]
			}
		}
		title = strings.TrimSpace(result["name"])
		id = result["id"]
	} else {
		title = base
	}
	line1 := fmt.Sprintf("<code>%s</code> %s", id, title)

	hashtag := "#gilticus"
	// timeLayout := time.RFC3339
	timeLayout := "2006-01-02 15:04:05"
	line2 := fmt.Sprintf("<code>%s</code> %s", file.ModTime.Format(timeLayout), hashtag)

	return line1 + "\n" + line2
}
