package send

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/meesooqa/files2tg/app/finder"
)

// mockSender implements TelegramSender и просто запоминает, что ему прислали
type mockSender struct {
	VideoSent *tb.Video
}

func (m *mockSender) Send(v tb.Video, bot *tb.Bot, rcp tb.Recipient, opts *tb.SendOptions) (*tb.Message, error) {
	m.VideoSent = &v
	return &tb.Message{Text: "ok"}, nil
}

func TestSend_SuccessVideo(t *testing.T) {
	sender := &mockSender{}
	client := TelegramClient{
		Opts:           &Options{Channel: "@channel"},
		Bot:            &tb.Bot{}, // non-nil, но для video‐send Bot не используется
		TelegramSender: sender,
		Formatter:      TelegramFormatter{}, // реальный форматтер (не вызывается при удачной video‐отправке)
	}

	file := finder.File{
		Name: "vid.mp4",
		Path: "/tmp/vid.mp4",
		Info: finder.VideoInfo{
			Width:    640,
			Height:   480,
			Duration: 7,
		},
	}

	require.NoError(t, client.Send(file))
	require.NotNil(t, sender.VideoSent, "должен был вызваться mockSender.Send")
	require.Equal(t, 640, sender.VideoSent.Width)
	require.Equal(t, 7, sender.VideoSent.Duration)
}

func TestSend_SkipIfBotNil(t *testing.T) {
	// если Bot==nil, Send просто возвращает nil без ошибок
	client := TelegramClient{
		Opts: &Options{Channel: "@x"},
		Bot:  nil,
	}
	require.NoError(t, client.Send(finder.File{Name: "any"}))
}

func TestSend_SkipIfChannelEmpty(t *testing.T) {
	// если Channel=="", тоже молча ноль
	client := TelegramClient{
		Opts: &Options{Channel: ""},
		Bot:  &tb.Bot{},
	}
	require.NoError(t, client.Send(finder.File{Name: "any"}))
}

func TestOptionsFromEnv(t *testing.T) {
	t.Setenv("TELEGRAM_CHAN", "@foo")
	t.Setenv("TELEGRAM_SERVER", "https://t.me")
	t.Setenv("TELEGRAM_TOKEN", "tok")
	t.Setenv("TELEGRAM_TIMEOUT", "5")

	opts := optionsFromEnv()
	require.Equal(t, "@foo", opts.Channel)
	require.Equal(t, "https://t.me", opts.Server)
	require.Equal(t, "tok", opts.Token)
	require.Equal(t, 5*time.Minute, opts.Timeout)
}

func TestNewTelegramClientFromEnv_NoToken(t *testing.T) {
	// когда переменная TELEGRAM_TOKEN пустая, бот не инициализируется (Bot=nil)
	t.Setenv("TELEGRAM_CHAN", "chan")
	t.Setenv("TELEGRAM_SERVER", "srv")
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("TELEGRAM_TIMEOUT", "2")

	cl, err := NewTelegramClientFromEnv()
	require.NoError(t, err)

	tc, ok := cl.(TelegramClient)
	require.True(t, ok)
	require.Nil(t, tc.Bot, "если токен пуст, Bot должен быть nil")
	require.Equal(t, 2*time.Minute, tc.Timeout)
}

func TestRecipient_Formatting(t *testing.T) {
	tests := []struct {
		raw, want string
	}{
		{"123456", "123456"},
		{"@me", "@me"},
		{"channel", "@channel"},
		{"-100789", "-100789"},
	}

	for _, tt := range tests {
		r := recipient{chatID: tt.raw}
		require.Equal(t, tt.want, r.Recipient(), "raw=%q", tt.raw)
	}
}
