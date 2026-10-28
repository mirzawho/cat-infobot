//go:build unix

package telegram

import (
	"errors"
	"os"
	"syscall"
)

var errFlockBusy = errors.New("lock held")

func tryFlock(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return errFlockBusy
	}
	return err
}

func funlock(f *os.File) error { return syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }
