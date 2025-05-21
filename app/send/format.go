package send

import (
	"path/filepath"
	"strings"

	"github.com/meesooqa/files2tg/app/finder"
)

type TelegramFormatter struct{}

// Format generates HTML message from provided finder.File
func (o *TelegramFormatter) Format(file finder.File) string {
	ext := filepath.Ext(file.Name)
	return strings.TrimSuffix(file.Name, ext)
}
