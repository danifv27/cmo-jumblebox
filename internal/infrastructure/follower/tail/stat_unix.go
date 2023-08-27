package tail

import (
	"fmt"
	"os"
	"syscall"
)

func stat(file *os.File) (*FileStat, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("follow: unexpected FileInfo.Sys() type. name %s, type %T", file.Name(), fi.Sys())
	}
	if sys == nil {
		return nil, fmt.Errorf("follow: FileInfo.Sys() returns nil. name %s", file.Name())
	}
	return &FileStat{Sys: *sys}, nil
}

// FileStat is a os specific file stat
type FileStat struct {
	Sys syscall.Stat_t
}

// porting from os.sameFile
func (s *FileStat) sameFile(other *FileStat) bool {

	return s.Sys.Dev == other.Sys.Dev && s.Sys.Ino == other.Sys.Ino
}

// Stat returns the named FileStat
func Stat(file *os.File) (*FileStat, error) {

	return stat(file)
}

// SameFile reports whether st1 and st2 represents the same file
func SameFile(st1, st2 *FileStat) bool {

	return st1.sameFile(st2)
}
