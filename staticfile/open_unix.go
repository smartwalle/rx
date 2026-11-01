//go:build unix

package staticfile

import (
	"os"
	"syscall"
)

func openFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
