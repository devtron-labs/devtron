package dir

import (
	"log"
	"os"
	"path"
)

func CheckOrCreateDevtronDir() (err error, devtronDirPath string) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("error occurred while finding home dir", "err", err)
		return err, ""
	}
	devtronDirPath = path.Join(userHomeDir, "./.devtron")
	err = os.MkdirAll(devtronDirPath, os.ModePerm)
	if err != nil {
		log.Fatalln("error occurred while creating folder", "path", devtronDirPath, "err", err)
		return err, ""
	}
	return err, devtronDirPath
}
