package send

import (
	"fmt"
	"time"

	"github.com/meesooqa/files2tg/app/finder"
)

type TelegramFormatter struct{}

// Format generates HTML message from provided finder.File
func (o *TelegramFormatter) Format(file finder.File) string {
	datetime := fmt.Sprintf("<code>%s</code>", file.ModTime.Format(time.RFC3339))
	return datetime + "\n" + file.Name
}
