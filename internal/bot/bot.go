// Package bot contains InfoBot's behaviour: the /start greeting and replying
// to messages with their technical details.
package bot

import (
	"context"
	"log/slog"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/mirza/infobot/internal/telegram"
)

// StartText is the greeting for /start.
const StartText = "سلام! 👋\n" +
	"من InfoBot هستم.\n" +
	"هر پیامی را برای من ارسال یا از یک چت دیگر Forward کن تا اطلاعات فنی و جزئیات قابل‌دسترسی آن را برایت نمایش بدهم."

// TextLimit is Telegram's maximum message length in characters.
const TextLimit = 4096

// Bot turns Bot API messages into replies.
type Bot struct {
	api *telegram.API
	log *slog.Logger
}

// New creates the bot logic. api must already be created for the same token.
func New(api *telegram.API, log *slog.Logger) *Bot {
	return &Bot{api: api, log: log}
}

// API returns the client the poller and the bot share.
func (b *Bot) API() *telegram.API { return b.api }

// HandleMessage answers one incoming message. Errors are logged by the poller.
func (b *Bot) HandleMessage(ctx context.Context, _ int64, m *telegram.Message) error {
	if m == nil {
		return nil
	}
	if isStart(m.Text) {
		return b.send(ctx, m.Chat.ID, m.MessageID, CommandReply(StartText))
	}
	for _, chunk := range Split(Render(*m), TextLimit) {
		if err := b.send(ctx, m.Chat.ID, m.MessageID, chunk); err != nil {
			return err
		}
	}
	return nil
}

// HandleEditedMessage answers an edited message with its current details.
func (b *Bot) HandleEditedMessage(ctx context.Context, updateID int64, m *telegram.Message) error {
	return b.HandleMessage(ctx, updateID, m)
}

func (b *Bot) send(ctx context.Context, chatID int64, replyTo int, text string) error {
	const attempts = 3
	var err error
	for i := 0; i < attempts; i++ {
		err = b.api.SendMessage(ctx, chatID, replyTo, text)
		if err == nil {
			return nil
		}
		if wait, ok := telegram.FloodWait(err); ok && wait <= time.Minute {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			continue
		}
		return err
	}
	return err
}

// isStart matches /start, /start@thisbot and /start with arguments.
func isStart(text string) bool {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}
	cmd, _, _ := strings.Cut(fields[0], "@")
	return strings.EqualFold(cmd, "/start")
}

// Split cuts text into chunks of at most max UTF-16 units, preferring line
// boundaries, so Telegram's 4096-character limit is never exceeded and no text
// is lost.
func Split(text string, max int) []string {
	if max <= 0 || max > TextLimit {
		max = TextLimit
	}
	if units(text) <= max {
		return []string{text}
	}
	var out []string
	var cur strings.Builder
	curLen := 0
	flush := func() {
		out = append(out, cur.String())
		cur.Reset()
		curLen = 0
	}
	add := func(line string) {
		if curLen > 0 {
			cur.WriteByte('\n')
			curLen++
		}
		cur.WriteString(line)
		curLen += units(line)
	}
	for _, line := range strings.Split(text, "\n") {
		for _, piece := range wrap(line, max) {
			if curLen > 0 && curLen+1+units(piece) > max {
				flush()
			}
			add(piece)
		}
	}
	flush()
	if len(out) == 0 {
		out = []string{""}
	}
	return out
}

// wrap breaks a single line into pieces of at most max units.
func wrap(line string, max int) []string {
	if units(line) <= max {
		return []string{line}
	}
	var out []string
	var cur strings.Builder
	n := 0
	for _, r := range line {
		u := 1
		if r > 0xFFFF {
			u = 2
		}
		if n+u > max {
			out = append(out, cur.String())
			cur.Reset()
			n = 0
		}
		cur.WriteRune(r)
		n += u
	}
	return append(out, cur.String())
}

func units(s string) int { return len(utf16.Encode([]rune(s))) }
