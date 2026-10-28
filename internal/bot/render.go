package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mirza/infobot/internal/telegram"
)

// Render returns the MarkdownV2 reply for one message: the message as
// indented JSON inside a code block. The payload is escaped, so any content is
// safe.
//
// Some Bot API servers and client versions omit the numeric identifiers that
// identify a forwarded message's source channel ("hidden" forwarding). When
// Telegram sent a source but no chat id, the reply says so explicitly instead
// of leaving the reader to guess.
func Render(m telegram.Message) string {
	note := forwardNote(m)
	if note == "" {
		return fence("json", PrettyJSON(m.Raw))
	}
	return telegram.EscapeMarkdown(note) + "\n" + fence("json", PrettyJSON(m.Raw))
}

// forwardNote describes forwarding information the JSON may not make obvious.
// It only states what Telegram actually sent; nothing is inferred.
func forwardNote(m telegram.Message) string {
	o := m.ForwardOrigin
	switch {
	case o == nil && m.ForwardFrom == nil && m.ForwardFromCh == nil && m.ForwardName == "" && m.ForwardDate == 0:
		return ""
	case o != nil && o.Chat != nil:
		return fmt.Sprintf("Forwarded from %s %q (chat_id: %d, source message_id: %d).",
			o.Chat.Type, chatTitle(*o.Chat), o.Chat.ID, o.MessageID)
	case o != nil && o.SenderUser != nil:
		return fmt.Sprintf("Forwarded from user %q (user_id: %d).", userName(*o.SenderUser), o.SenderUser.ID)
	case o != nil && o.Type == "hidden_user":
		return fmt.Sprintf("Forwarded from a hidden source whose name Telegram sent as %q; no chat_id or user_id is available.", o.SenderUserName)
	case m.ForwardFromCh != nil:
		return fmt.Sprintf("Forwarded from %s %q (chat_id: %d).", m.ForwardFromCh.Type, chatTitle(*m.ForwardFromCh), m.ForwardFromCh.ID)
	case m.ForwardFrom != nil:
		return fmt.Sprintf("Forwarded from user %q (user_id: %d).", userName(*m.ForwardFrom), m.ForwardFrom.ID)
	case m.ForwardName != "":
		return fmt.Sprintf("Forwarded from a hidden source whose name Telegram sent as %q; no chat_id or user_id is available.", m.ForwardName)
	}
	return "Telegram marked this message as forwarded but did not send the source chat/user id."
}

func chatTitle(c telegram.Chat) string {
	if c.Title != "" {
		return c.Title
	}
	return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

func userName(u telegram.User) string {
	if n := strings.TrimSpace(u.FirstName + " " + u.LastName); n != "" {
		return n
	}
	return u.Username
}

// PrettyJSON re-indents raw JSON with two spaces.
func PrettyJSON(raw []byte) string {
	if len(raw) == 0 {
		return "{}"
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}

// fence wraps text in a MarkdownV2 code block. Inside a pre block only ` and \
// need escaping, and no run of backticks may close it.
func fence(lang, text string) string {
	return "```" + lang + "\n" + escapeCode(text) + "\n```"
}

func escapeCode(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "`", "\\`")
	return r.Replace(s)
}

// CommandReply renders a short reply for a command.
func CommandReply(text string) string {
	return "*InfoBot*\n" + telegram.EscapeMarkdown(text)
}
