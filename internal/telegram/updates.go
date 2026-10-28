package telegram

import (
	"context"
	"net/url"
	"strconv"
)

// Update carries one new or edited message.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
	Edited   *Message `json:"edited_message"`
}

// GetUpdates performs one long-poll request. Telegram holds the request for
// up to timeout seconds when there is nothing new.
func (a *API) GetUpdates(ctx context.Context, offset int64, timeoutSeconds int) ([]Update, error) {
	p := url.Values{}
	p.Set("offset", strconv.FormatInt(offset, 10))
	p.Set("limit", "100")
	p.Set("timeout", strconv.Itoa(timeoutSeconds))
	p.Set("allowed_updates", `["message","edited_message"]`)
	return call[[]Update](ctx, a, "getUpdates", p)
}
