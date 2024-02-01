package util

import (
	"os"
	"strconv"
)

func MoveFileToDestination(filePath, destinationPath string) error {
	err := os.Rename(filePath, destinationPath)
	if err != nil {
		return err
	}
	return nil
}

func ConvertIntArrayToStringArray(req []int) []string {
	var resp []string
	for _, item := range req {
		resp = append(resp, strconv.Itoa(item))
	}
	return resp
}
