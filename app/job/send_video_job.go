package job

import (
	"fmt"

	"github.com/meesooqa/files2tg/app/finder"
	"github.com/meesooqa/files2tg/app/send"
)

// SendVideoJob send finder.File to Telegram
type SendVideoJob struct {
	BaseJob
	File           finder.File
	Stars          int
	TelegramClient send.Client
}

// Execute implements SendVideoJob
func (o SendVideoJob) Execute() error {
	fmt.Printf("Start processing file: %s\n", o.File.Name)
	if err := o.TelegramClient.Send(o.File, o.Stars); err != nil {
		return fmt.Errorf("failed to send to Telegram: %v", err)
	}
	return nil
}
