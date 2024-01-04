package bean

import "errors"

const (
	YamlSeparator string = "---\n"
)

var ErrNoCommitFound = errors.New("no commit found")
