package util

import (
	"os"
	"strconv"
	"strings"
)

const RELEASE_NOT_EXIST = "release not exist"
const NOT_FOUND = "not found"

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

func CheckAppReleaseNotExist(err error) bool {
	// RELEASE_NOT_EXIST check for helm App and NOT_FOUND check for argo app
	return strings.Contains(err.Error(), NOT_FOUND) || strings.Contains(err.Error(), RELEASE_NOT_EXIST)
}
