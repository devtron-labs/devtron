package file

import (
	"os"

	"github.com/pkg/errors"
)

// IsDirectory returns whether or not the given file is a directory
func IsDirectory(path string) (bool, error) {
	fileOrDir, err := os.Open(path)
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer func() { _ = fileOrDir.Close() }()
	stat, err := fileOrDir.Stat()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return stat.IsDir(), nil
}

// Exists returns whether or not a path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
