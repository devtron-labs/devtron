package helper

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"strings"
)

func FilterReservedPathFromOutputDirPath(outputDirectoryPath []string) []string {
	var newOutputDirPath []string
	for _, path := range outputDirectoryPath {
		if !strings.HasPrefix(path, constants.CiRunnerWorkingDir) {
			newOutputDirPath = append(newOutputDirPath, path)
		}
	}
	return newOutputDirPath
}
