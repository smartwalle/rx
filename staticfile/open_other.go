//go:build !unix

package staticfile

import "os"

func openFile(name string) (*os.File, error) {
	return os.Open(name)
}
