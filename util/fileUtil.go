package util

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"
)

func CreateFileWithData(folderPath, fileName, content string) (string, error) {
	filePath := path.Join(folderPath, fileName)
	// if file exists then delete file
	if _, err := os.Stat(filePath); os.IsExist(err) {
		os.Remove(filePath)
	} else if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return "", err
		}
	}
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err2 := f.WriteString(content)
	if err2 != nil {
		return "", err
	}
	return filePath, nil
}

func DeleteFolder(folderPath string) error {
	err := os.RemoveAll(folderPath)
	if err != nil {
		fmt.Println("Error deleting folder:", err)
	} else {
		fmt.Println("Folder deleted successfully")
	}
	return nil
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func GetRandomName() string {
	r1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	randomName := fmt.Sprintf(strconv.FormatInt(r1, 10))
	return randomName
}
