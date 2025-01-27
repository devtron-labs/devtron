package utils

import (
	"errors"
	"fmt"
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

const (
	PermissionMode = 0644
)

func CreateDirectory(path string) error {
	err := os.MkdirAll(path, PermissionMode)
	if err != nil {
		fmt.Println("error in creating directory", "err", err)
		return err
	}
	return nil
}

func CheckFileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err == nil {
		// exists
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		// not exists
		return false, nil
	} else {
		// Some other error
		return false, err
	}
}

func WriteToFile(file string, fileName string) error {
	err := os.WriteFile(fileName, []byte(file), PermissionMode)
	if err != nil {
		fmt.Println("error in writing results to json file", "err", err)
		return err
	}
	return nil
}
