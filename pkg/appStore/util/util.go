package util

import (
	"os"
)

func MoveFileToDestination(filePath, destinationPath string) error {
	err := os.Rename(filePath, destinationPath)
	if err != nil {
		return err
	}
	return nil
}

func CreateFileAtFilePathAndWrite(filePath, fileContent string) (string, error) {
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return filePath, err
	}
	_, err = file.Write([]byte(fileContent))
	if err != nil {
		return filePath, err
	}
	return filePath, err
}
