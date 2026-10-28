package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Handler receives every new (or edited) message. updateID is already
// consumed by the poller once the handler returns (or fails).
type Handler func(ctx context.Context, updateID int64, m *Message) error

// Poller is the long-polling loop over the Bot API.
type Poller struct {
	api            *API
	log            *slog.Logger
	onMessage      Handler
	onEdited       Handler
	requestTimeout int // seconds Telegram keeps the getUpdates request open
}

// NewPoller creates a poller. Messages and edited messages share one offset.
func NewPoller(cfg Config, log *slog.Logger, onMessage, onEdited Handler) *Poller {
	return &Poller{
		api:            NewAPI(cfg.Token, cfg.Endpoint),
		log:            log,
		onMessage:      onMessage,
		onEdited:       onEdited,
		requestTimeout: 25,
	}
}

// Run blocks until ctx is cancelled or the token is rejected. Transient
// network and rate-limit errors are retried with backoff; they never crash
// the bot.
func (p *Poller) Run(ctx context.Context) error {
	me, err := p.api.GetMe(ctx)
	if err != nil {
		var e *APIError
		if errors.As(err, &e) && e.Code == 401 {
			return fmt.Errorf("Telegram rejected BOT_TOKEN: %w", err)
		}
		return fmt.Errorf("getMe: %w", err)
	}
	p.log.Info("connected", "bot", me.Username, "bot_id", me.ID)

	var offset int64
	backoff := time.Second
	for {
		updates, err := p.api.GetUpdates(ctx, offset, p.requestTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if wait, flood := FloodWait(err); flood {
				p.log.Warn("rate limited", "retry_after", wait)
				if !sleep(ctx, wait) {
					return nil
				}
				continue
			}
			var e *APIError
			if errors.As(err, &e) && e.Code >= 400 && e.Code < 500 && e.Code != 409 {
				return fmt.Errorf("getUpdates rejected: %w", err)
			}
			p.log.Warn("poll failed, retrying", "error", err, "backoff", backoff)
			if !sleep(ctx, backoff) {
				return nil
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second
		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}
			if err := p.handle(ctx, u); err != nil {
				p.log.Error("handle update failed", "update_id", u.UpdateID, "error", err)
			}
		}
	}
}

func (p *Poller) handle(ctx context.Context, u Update) error {
	switch {
	case u.Message != nil:
		if p.onMessage == nil {
			return nil
		}
		return p.onMessage(ctx, u.UpdateID, u.Message)
	case u.Edited != nil:
		if p.onEdited == nil {
			return nil
		}
		return p.onEdited(ctx, u.UpdateID, u.Edited)
	}
	return nil
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
