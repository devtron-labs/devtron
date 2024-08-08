package utils

import (
	"io/ioutil"
	"os"
	"path"
)

func CreateFolderAndFileWithContent(content string, fileName string, folderPath string) (string, error) {
	tlsFilePath := path.Join(folderPath, fileName)

	// if file exists then delete file
	if _, err := os.Stat(tlsFilePath); os.IsExist(err) {
		os.Remove(tlsFilePath)
	}
	// create dirs
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		return "", err
	}

	// create file with content
	err = ioutil.WriteFile(tlsFilePath, []byte(content), 0600)
	if err != nil {
		return "", err
	}
	return tlsFilePath, nil
}

func DeleteAFileIfExists(path string) error {
	if _, err := os.Open(path); err == nil {
		err = os.Remove(path)
		return err
	}
	return nil
}
