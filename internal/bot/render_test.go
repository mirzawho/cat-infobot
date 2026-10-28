package bot

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/mirza/infobot/internal/telegram"
)

var sampleJSON = `{
  "message_id": 3,
  "date": 1761577325,
  "from": {"id": 1780243209, "is_bot": false, "first_name": "Mirza", "username": "mirza"},
  "chat": {"id": 1780243209, "type": "private", "first_name": "Mirza", "username": "mirza"},
  "text": "سلام _ * [x](y) backtick"
}`

func message(t *testing.T, js string) telegram.Message {
	t.Helper()
	var m telegram.Message
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestRenderJSON(t *testing.T) {
	out := Render(message(t, sampleJSON))
	if !strings.HasPrefix(out, "```json\n") {
		t.Fatalf("unexpected opening fence:\n%s", out)
	}
	if !strings.HasSuffix(out, "\n```") {
		t.Fatalf("missing closing fence:\n%s", out)
	}
	body := strings.TrimSuffix(strings.TrimPrefix(out, "```json\n"), "\n```")
	// The code block protects the payload; unescaping must return valid JSON.
	var decoded map[string]any
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("JSON inside the code block is not valid: %v\n%s", err, body)
	}
	if decoded["message_id"].(float64) != 3 {
		t.Errorf("message_id = %v", decoded["message_id"])
	}
	if !strings.Contains(body, "\n  \"date\": 1761577325,") {
		t.Errorf("JSON is not indented:\n%s", body)
	}
}

func TestRenderEscapesBackticksInPayload(t *testing.T) {
	m := message(t, `{"message_id": 1, "chat": {"id": 1, "type": "private"}, "text": "a`+"```"+`b"}`)
	out := Render(m)
	// Only the fences may contain three backticks in a row.
	if strings.Count(out, "```") != 2 {
		t.Errorf("code block broken by payload:\n%s", out)
	}
}

func TestRenderEmptyRaw(t *testing.T) {
	if got := PrettyJSON(nil); got != "{}" {
		t.Errorf("PrettyJSON(nil) = %q", got)
	}
}

func TestEscapeMarkdown(t *testing.T) {
	got := telegram.EscapeMarkdown("a_b*c[d](e)~f`g>h#i+j-k=l|m{n}o.p!q\\")
	want := `a\_b\*c\[d\]\(e\)\~f\` + "`" + `g\>h\#i\+j\-k\=l\|m\{n\}o\.p\!q\\`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
	if plain := telegram.EscapeMarkdown("abc 123 سلام"); plain != "abc 123 سلام" {
		t.Errorf("escaped harmless text: %q", plain)
	}
}

func TestSplitKeepsFencesClosed(t *testing.T) {
	// A message whose JSON exceeds one Telegram message: the splitter must not
	// cut inside a line.
	big := strings.Repeat("x", 5000*4)
	m := message(t, `{"message_id": 1, "chat": {"id": 1, "type": "private"}, "text": "`+big+`"}`)
	chunks := Split(Render(m), 4096)
	if len(chunks) < 2 {
		t.Fatalf("expected split, got %d chunk(s)", len(chunks))
	}
	for i, c := range chunks {
		if n := len(utf16.Encode([]rune(c))); n > 4096 {
			t.Errorf("chunk %d has %d units", i, n)
		}
	}
}

func TestStartMatcher(t *testing.T) {
	for _, text := range []string{"/start", "/start@infobot", "/Start arg"} {
		if !isStart(text) {
			t.Errorf("%q should be /start", text)
		}
	}
	for _, text := range []string{"/starter", "start", ""} {
		if isStart(text) {
			t.Errorf("%q should not be /start", text)
		}
	}
}

func TestForwardSourceLine(t *testing.T) {
	out := Render(message(t, `{
	  "message_id": 1, "chat": {"id": 1, "type": "private"}, "text": "x",
	  "forward_origin": {"type": "channel", "date": 1,
	    "chat": {"id": -100123, "type": "channel", "title": "Source Channel"},
	    "message_id": 77}
	}`))
	if !strings.HasPrefix(out, "Forwarded from channel \"Source Channel\" \\(chat\\_id: \\-100123, source message\\_id: 77\\)\\.\n```json\n") {
		t.Fatalf("unexpected forward line:\n%s", out)
	}

	out = Render(message(t, `{
	  "message_id": 1, "chat": {"id": 1, "type": "private"}, "text": "x",
	  "forward_origin": {"type": "hidden_user", "date": 1, "sender_user_name": "Someone"}
	}`))
	if !strings.Contains(out, "hidden source whose name Telegram sent as \"Someone\"") {
		t.Fatalf("hidden source not explained:\n%s", out)
	}

	out = Render(message(t, `{"message_id": 1, "chat": {"id": 1, "type": "private"}, "text": "x"}`))
	if strings.HasPrefix(out, "Forwarded") {
		t.Fatalf("plain message labelled as forward:\n%s", out)
	}
}
