package commandManager

import (
	"github.com/go-git/go-billy/v5/osfs"
	"os"
	"path/filepath"
)

func LocateGitRepo(path string) error {

	if _, err := filepath.Abs(path); err != nil {
		return err
	}
	fst := osfs.New(path)
	_, err := fst.Stat(".git")
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}
