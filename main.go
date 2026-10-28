// Command infobot reports the technical details of any message sent to it.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mirza/infobot/internal/bot"
	"github.com/mirza/infobot/internal/telegram"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()}))

	cfg, err := telegram.LoadConfig()
	if err != nil {
		logger.Error("configuration error", "error", err)
		return 1
	}

	lock, err := telegram.AcquireLock(telegram.DataDir())
	if err != nil {
		if errors.Is(err, telegram.ErrLocked) {
			logger.Error("another infobot instance is already polling; refusing to start")
		} else {
			logger.Error("cannot acquire instance lock", "error", err)
		}
		return 1
	}
	defer lock.Release()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api := telegram.NewAPI(cfg.Token, cfg.Endpoint)
	if debugAPI() {
		api.Debug(logger)
	}
	b := bot.New(api, logger)
	poller := telegram.NewPoller(cfg, logger, b.HandleMessage, b.HandleEditedMessage)

	logger.Info("starting infobot")
	if err := poller.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("bot stopped with error", "error", err)
		return 1
	}
	logger.Info("shutdown complete")
	return 0
}

// logLevel reads INFOBOT_LOG_LEVEL (debug, info, warn, error).
func logLevel() slog.Level {
	switch os.Getenv("INFOBOT_LOG_LEVEL") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// debugAPI reports whether INFOBOT_DEBUG_API=1 was set.
func debugAPI() bool {
	switch os.Getenv("INFOBOT_DEBUG_API") {
	case "1", "true", "yes":
		return true
	}
	return false
}
