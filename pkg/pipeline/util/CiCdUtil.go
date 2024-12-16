package util

import (
	"github.com/devtron-labs/devtron/util/urlUtil"
)

func IsValidUrlSubPath(subPath string) bool {
	url := "http://127.0.0.1:8080/" + subPath
	return urlUtil.IsValidUrl(url)
}
