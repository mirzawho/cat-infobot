package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIError is a rejection returned by the Bot API.
type APIError struct {
	Method      string
	Code        int
	Description string
	RetryAfter  int
}

func (e *APIError) Error() string { return fmt.Sprintf("%s: %d %s", e.Method, e.Code, e.Description) }

// API is the HTTP Bot API client.
type API struct {
	token  string
	base   string
	http   *http.Client
	debug  bool
	logger *slog.Logger
}

// NewAPI creates a client; an empty endpoint means api.telegram.org.
// SetDebug enables request/response logging (tokens are never logged).
func NewAPI(token, endpoint string) *API {
	base := strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		base = "https://api.telegram.org"
	}
	return &API{token: token, base: base, http: &http.Client{Timeout: 45 * time.Second}}
}

// Debug logs every outgoing request body and every response body.
func (a *API) Debug(logger *slog.Logger) {
	a.debug = true
	a.logger = logger
}

func (a *API) logDebug(msg string, fields ...any) {
	if a.debug && a.logger != nil {
		a.logger.Debug(msg, fields...)
	}
}

type envelope[T any] struct {
	OK          bool   `json:"ok"`
	Result      T      `json:"result"`
	Description string `json:"description"`
	ErrorCode   int    `json:"error_code"`
	Parameters  *struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

func call[T any](ctx context.Context, api *API, method string, params url.Values) (T, error) {
	var zero T
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		api.base+"/bot"+api.token+"/"+method, strings.NewReader(params.Encode()))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	api.logDebug("api request", "method", method, "form", params.Encode())
	resp, err := api.http.Do(req)
	if err != nil {
		return zero, fmt.Errorf("%s: %w", method, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return zero, fmt.Errorf("%s: read response: %w", method, err)
	}
	api.logDebug("api response", "method", method, "body", truncate(string(raw), 4000))
	var env envelope[T]
	if err := json.Unmarshal(raw, &env); err != nil {
		return zero, fmt.Errorf("%s: decode response: %w", method, err)
	}
	if !env.OK {
		e := &APIError{Method: method, Code: env.ErrorCode, Description: env.Description}
		if env.Parameters != nil {
			e.RetryAfter = env.Parameters.RetryAfter
		}
		return zero, e
	}
	return env.Result, nil
}

type apiUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// GetMe validates the token and returns the bot account.
func (a *API) GetMe(ctx context.Context) (apiUser, error) {
	return call[apiUser](ctx, a, "getMe", url.Values{})
}

// MarkdownMode is the Bot API parse mode the bot uses; every payload escapes
// user text with EscapeMarkdown.
const MarkdownMode = "MarkdownV2"

// SendMessage posts MarkdownV2 text to a chat, as a reply to replyTo when
// non-zero.
func (a *API) SendMessage(ctx context.Context, chatID int64, replyTo int, text string) error {
	params := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"parse_mode":               MarkdownMode,
		"disable_web_page_preview": true,
	}
	if replyTo != 0 {
		// Sent as JSON so the nested object arrives as an object; an
		// url.Values string would be percent-encoded and ignored.
		params["reply_parameters"] = map[string]any{
			"message_id":                  replyTo,
			"allow_sending_without_reply": true,
		}
	}
	_, err := callJSON[json.RawMessage](ctx, a, "sendMessage", params)
	return err
}

// callJSON posts a JSON body and decodes the Bot API envelope.
func callJSON[T any](ctx context.Context, api *API, method string, params map[string]any) (T, error) {
	var zero T
	body, err := json.Marshal(params)
	if err != nil {
		return zero, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		api.base+"/bot"+api.token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")
	api.logDebug("api request", "method", method, "body", truncate(string(body), 4000))
	return decodeEnvelope[T](api, req)
}

func decodeEnvelope[T any](api *API, req *http.Request) (T, error) {
	var zero T
	resp, err := api.http.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if readErr != nil {
		return zero, readErr
	}
	api.logDebug("api response", "request", truncate(req.URL.Path, 200), "body", truncate(string(raw), 4000))
	var env envelope[T]
	if err := json.Unmarshal(raw, &env); err != nil {
		return zero, fmt.Errorf("%w", err)
	}
	if !env.OK {
		e := &APIError{Code: env.ErrorCode, Description: env.Description}
		if env.Parameters != nil {
			e.RetryAfter = env.Parameters.RetryAfter
		}
		return zero, e
	}
	return env.Result, nil
}

// EscapeMarkdown escapes a literal for MarkdownV2. It is applied to every
// piece of user-controlled text (message text, names, JSON payloads).
func EscapeMarkdown(s string) string {
	var b strings.Builder
	b.Grow(len(s) + len(s)/4)
	for _, r := range s {
		if strings.ContainsRune("\\_*[]()~`>#+-=|{}.!", r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FloodWait reports the retry delay of a rate-limit error.
func FloodWait(err error) (time.Duration, bool) {
	var e *APIError
	if errors.As(err, &e) && e.Code == 429 && e.RetryAfter > 0 {
		return time.Duration(e.RetryAfter) * time.Second, true
	}
	return 0, false
}

// truncate shortens long bodies for logs.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
