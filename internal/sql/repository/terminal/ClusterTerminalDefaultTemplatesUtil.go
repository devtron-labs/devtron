package terminal

import (
	"embed"
	"fmt"
)

//go:embed static/*
var staticTemplates embed.FS

func GetDefaultTerminalAccessServiceAccount() string {
	return readContent("static/TerminalAccessServiceAccount")
}

func GetDefaultTerminalAccessPodTemplate() string {
	return readContent("static/TerminalAccessPodTemplate")
}

func GetDefaultTerminalAccessRoleBindingTemplate() string {
	return readContent("static/TerminalAccessRoleBinding")
}

func readContent(fileName string) string {
	//filePath := path.Join(basePath, fileName)
	fileContent, err := staticTemplates.ReadFile(fileName)
	if err != nil {
		fmt.Println("error occurred while reading json file", "fileName", fileName, "err", err)
	}
	return string(fileContent)
}
