//go:build linux || freebsd || darwin
// +build linux freebsd darwin

package tail

import "os"

// Open opens the named file for reading and following.
func OpenFile(name string) (*os.File, error) {

	return os.OpenFile(name, os.O_RDONLY, 0)
}
