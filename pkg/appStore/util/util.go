package util

import "os"

func MoveFileToDestination(filePath, destinationPath string) error {
	err := os.Rename(filePath, destinationPath)
	if err != nil {
		return err
	}
	return nil
}
