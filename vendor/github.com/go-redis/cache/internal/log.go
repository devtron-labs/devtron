package internal

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

var (
	// Log is the instance of a Logger interface that cache writes errors to.
	Log Logger = log.New(os.Stderr, "", log.LstdFlags)
)
