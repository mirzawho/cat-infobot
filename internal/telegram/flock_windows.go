//go:build windows

package telegram

import (
	"errors"
	"os"
)

var errFlockBusy = errors.New("lock held")

// Windows builds are for local development only; a second instance is allowed
// to start, the Bot API getUpdates offset keeps replies from duplicating.
func tryFlock(*os.File) error { return nil }

func funlock(*os.File) error { return nil }
