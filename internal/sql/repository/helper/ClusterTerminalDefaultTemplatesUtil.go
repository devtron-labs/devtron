package helper

import (
	"fmt"
	"os"
)

func GetDefaultTerminalAccessServiceAccount() string {
	return readContent("TerminalAccessServiceAccount")
}

func GetDefaultTerminalAccessPodTemplate() string {
	return readContent("TerminalAccessPodTemplate")
}

func GetDefaultTerminalAccessRoleBindingTemplate() string {
	return readContent("TerminalAccessRoleBinding")
}

func readContent(fileName string) string {
	//filePath := path.Join(basePath, fileName)
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("error occurred while reading json file", "fileName", fileName, "err", err)
	}
	return string(fileContent)
}
