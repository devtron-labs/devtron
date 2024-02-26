package models

import "strings"

type NamespaceNotExistError struct {
	Err error
}

func (err NamespaceNotExistError) Error() string {
	return err.Err.Error()
}

func (err *NamespaceNotExistError) Unwrap() error {
	return err.Err
}

func IsErrorWhileGeneratingManifest(err error) bool {
	if strings.Contains(err.Error(), "error converting YAML to JSON") || strings.Contains(err.Error(), "error occured while generating manifest") {
		return true
	}
	return false
}
