package send

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	tb "gopkg.in/telebot.v4"

	"github.com/meesooqa/files2tg/app/finder"
)

type Options struct {
	Channel string
	Server  string
	Token   string
	Timeout time.Duration
}

type Client interface {
	Send(file finder.File) error
}

type ClientFactory interface {
	NewClient() (Client, error)
}

type EnvClientFactory struct{}

func (f *EnvClientFactory) NewClient() (Client, error) {
	return NewTelegramClientFromEnv()
}

// TelegramSender is the interface for sending messages to telegram
type TelegramSender interface {
	Send(tb.Video, *tb.Bot, tb.Recipient, *tb.SendOptions) (*tb.Message, error)
	SendPaid(tb.Video, *tb.Bot, tb.Recipient, *tb.SendOptions) (*tb.Message, error)
}

type TelegramClient struct {
	Opts           *Options
	Bot            *tb.Bot
	Timeout        time.Duration
	TelegramSender TelegramSender
	Formatter      TelegramFormatter
}

func optionsFromEnv() *Options {
	telegramTimeout, _ := strconv.Atoi(os.Getenv("TELEGRAM_TIMEOUT"))
	return &Options{
		Channel: os.Getenv("TELEGRAM_CHAN"),
		Server:  os.Getenv("TELEGRAM_SERVER"),
		Token:   os.Getenv("TELEGRAM_TOKEN"),
		Timeout: time.Duration(telegramTimeout) * time.Minute,
	}
}

// NewTelegramClientFromEnv init telegram client from ENV
func NewTelegramClientFromEnv() (client Client, err error) {
	opts := optionsFromEnv()
	client, err = newTelegramClient(
		opts,
		&TelegramSenderImpl{},
		TelegramFormatter{},
	)
	return
}

// newTelegramClient init telegram client
func newTelegramClient(opts *Options, tgs TelegramSender, tf TelegramFormatter) (Client, error) {
	token := strings.TrimSpace(opts.Token)
	apiURL := strings.TrimSpace(opts.Server)
	timeout := opts.Timeout
	log.Printf("[INFO] create telegram client for %s, timeout: %s", apiURL, timeout)
	if timeout == 0 {
		timeout = time.Second * 60
	}

	if token == "" {
		return TelegramClient{Opts: opts, Bot: nil, Timeout: timeout}, nil
	}

	bot, err := tb.NewBot(tb.Settings{
		URL:    apiURL,
		Token:  token,
		Client: &http.Client{Timeout: timeout},
	})
	if err != nil {
		return TelegramClient{}, err
	}

	result := TelegramClient{
		Opts:           opts,
		Bot:            bot,
		Timeout:        timeout,
		TelegramSender: tgs,
		Formatter:      tf,
	}
	return result, err
}

func (o TelegramClient) Send(file finder.File) (err error) {
	channelID := o.Opts.Channel
	if o.Bot == nil || channelID == "" {
		return nil
	}

	message, err := o.sendVideo(channelID, file)
	if err != nil && strings.Contains(err.Error(), "Request Entity Too Large") {
		message, err = o.sendText(channelID, file)
	}

	if err != nil {
		return errors.Wrapf(err, "can't send to telegram for %+v", file.Name)
	}

	log.Printf("[DEBUG] telegram message sent: \n%s", message.Text)
	//log.Printf("[DEBUG] telegram message sent: \n%s", message.Text, message.Caption)
	return nil
}

func (o TelegramClient) sendText(channelID string, file finder.File) (*tb.Message, error) {
	message, err := o.Bot.Send(
		recipient{chatID: channelID},
		o.Formatter.Format(file),
		tb.ModeHTML,
		tb.NoPreview,
	)

	return message, err
}

func (o TelegramClient) sendVideo(channelID string, file finder.File) (*tb.Message, error) {
	// TODO defer os.Remove(file.Path)

	attachment := tb.Video{
		File:     tb.FromDisk(file.Path),
		Width:    file.Info.Width,
		Height:   file.Info.Height,
		Duration: file.Info.Duration,

		Streaming: true,
		Caption:   o.getMessageHTML(file),
	}
	return o.TelegramSender.SendPaid(attachment, o.Bot, recipient{chatID: channelID}, &tb.SendOptions{ParseMode: tb.ModeHTML})
}

// getMessageHTML generates HTML message from provided media.Info
func (o TelegramClient) getMessageHTML(file finder.File) string {
	return o.Formatter.Format(file)
}

type recipient struct {
	chatID string
}

func (r recipient) Recipient() string {
	if _, err := strconv.ParseInt(r.chatID, 10, 64); err != nil && !strings.HasPrefix(r.chatID, "@") {
		return "@" + r.chatID
	}

	return r.chatID
}

// TelegramSenderImpl is a TelegramSender implementation that sends messages to Telegram for real
type TelegramSenderImpl struct{}

// Send sends a message to Telegram
func (tg *TelegramSenderImpl) Send(attachment tb.Video, bot *tb.Bot, rcp tb.Recipient, opts *tb.SendOptions) (*tb.Message, error) {
	return attachment.Send(bot, rcp, opts)
}

// SendPaid sends a paid message to Telegram
func (tg *TelegramSenderImpl) SendPaid(attachment tb.Video, bot *tb.Bot, rcp tb.Recipient, opts *tb.SendOptions) (*tb.Message, error) {
	// TODO stars 1000
	return bot.SendPaid(rcp, 1000, tb.PaidAlbum{&attachment}, opts)
}
