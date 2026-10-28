package telegram

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Lock is the single-instance lock every bot process must hold before it
// starts polling, so two copies never answer the same update twice.
type Lock struct{ f *os.File }

// ErrLocked means another instance already holds the lock.
var ErrLocked = errors.New("another infobot instance is already running")

// AcquireLock takes <dir>/infobot.lock without blocking. The kernel drops it
// when the process exits, so a crash never leaves a stale lock.
func AcquireLock(dir string) (*Lock, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create lock directory: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(dir, "infobot.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := tryFlock(f); err != nil {
		_ = f.Close()
		if errors.Is(err, errFlockBusy) {
			return nil, ErrLocked
		}
		return nil, err
	}
	return &Lock{f: f}, nil
}

// Release drops the lock.
func (l *Lock) Release() {
	if l == nil || l.f == nil {
		return
	}
	_ = funlock(l.f)
	_ = l.f.Close()
	l.f = nil
}
