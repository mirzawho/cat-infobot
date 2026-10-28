package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// captureAPI returns an API whose HTTP transport records the sendMessage body
// (no sockets are opened).
func captureAPI(t *testing.T, response string) (*API, *captured) {
	t.Helper()
	c := &captured{}
	api := NewAPI("123:abc", "https://bot-api.test")
	api.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		c.method = r.Method
		c.path = r.URL.Path
		c.contentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		c.body = body
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(response)),
		}, nil
	})
	return api, c
}

type captured struct {
	method, path, contentType string
	body                      []byte
}

func (c *captured) json(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(c.body, &m); err != nil {
		t.Fatalf("request body is not JSON: %v (%s)", err, c.body)
	}
	return m
}

func TestSendMessageReplies(t *testing.T) {
	api, c := captureAPI(t, `{"ok":true,"result":{"message_id":1}}`)
	if err := api.SendMessage(context.Background(), 42, 7, "```json\n{}\n```"); err != nil {
		t.Fatal(err)
	}
	if c.method != http.MethodPost || c.path != "/bot123:abc/sendMessage" {
		t.Errorf("request = %s %s", c.method, c.path)
	}
	if c.contentType != "application/json" {
		t.Errorf("content type = %q", c.contentType)
	}
	got := c.json(t)
	if got["chat_id"].(float64) != 42 {
		t.Errorf("chat_id = %v", got["chat_id"])
	}
	if got["parse_mode"] != "MarkdownV2" {
		t.Errorf("parse_mode = %v", got["parse_mode"])
	}
	rp, ok := got["reply_parameters"].(map[string]any)
	if !ok {
		t.Fatalf("reply_parameters is not an object: %T %v", got["reply_parameters"], got["reply_parameters"])
	}
	if rp["message_id"].(float64) != 7 || rp["allow_sending_without_reply"] != true {
		t.Errorf("reply_parameters = %v", rp)
	}
}

func TestSendMessageWithoutReply(t *testing.T) {
	api, c := captureAPI(t, `{"ok":true,"result":{"message_id":1}}`)
	if err := api.SendMessage(context.Background(), 1, 0, "x"); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.json(t)["reply_parameters"]; ok {
		t.Errorf("reply_parameters sent for replyTo=0")
	}
}

func TestSendMessageAPIError(t *testing.T) {
	api, _ := captureAPI(t, `{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":3}}`)
	err := api.SendMessage(context.Background(), 1, 2, "x")
	wait, ok := FloodWait(err)
	if !ok || wait.Seconds() != 3 {
		t.Fatalf("FloodWait(%v) = %v, %v", err, wait, ok)
	}
}

func TestGetUpdatesForm(t *testing.T) {
	api, c := captureAPI(t, `{"ok":true,"result":[{"update_id":5,"message":{"message_id":3,"chat":{"id":1,"type":"private"},"text":"hi"}}]}`)
	updates, err := api.GetUpdates(context.Background(), 3, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 1 || updates[0].Message == nil || updates[0].Message.Text != "hi" {
		t.Fatalf("updates = %+v", updates)
	}
	if c.contentType != "application/x-www-form-urlencoded" {
		t.Errorf("content type = %q", c.contentType)
	}
	form, err := url.ParseQuery(string(c.body))
	if err != nil {
		t.Fatal(err)
	}
	if form.Get("offset") != "3" || form.Get("timeout") != "25" {
		t.Errorf("form = %v", form)
	}
	if string(updates[0].Message.Raw) == "" {
		t.Error("raw JSON not captured")
	}
}
