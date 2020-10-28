package gocd

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Set logging level and type constants
const (
	LogLevelEnvVarName = "GOCD_LOG_LEVEL"
	LogLevelDefault    = "WARNING"
	LogTypeEnvVarName  = "GOCD_LOG_TYPE"
	LogTypeDefault     = "TEXT"
)

var logLevels = map[string]logrus.Level{
	"PANIC":   logrus.PanicLevel,
	"FATAL":   logrus.FatalLevel,
	"ERROR":   logrus.ErrorLevel,
	"WARNING": logrus.WarnLevel,
	"INFO":    logrus.InfoLevel,
	"DEBUG":   logrus.DebugLevel,
}

var logFormat = map[string]logrus.Formatter{
	"JSON": &logrus.JSONFormatter{},
	"TEXT": &logrus.TextFormatter{},
}

// SetupLogging based on Environment Variables
//
//  Set Logging level with $GOCD_LOG_LEVEL
//  Allowed Values:
//    - DEBUG
//    - INFO
//    - WARNING
//    - ERROR
//    - FATAL
//    - PANIC
//
//  Set Logging type  with $GOCD_LOG_TYPE
//  Allowed Values:
//    - JSON
//    - TEXT
func SetupLogging(log *logrus.Logger) {
	log.SetLevel(logLevels[getLogLevel()])

	log.Formatter = logFormat[getLogType()]
}

// Get the log type from env variables
func getLogType() (logType string) {
	logType = os.Getenv(LogTypeEnvVarName)
	if len(logType) == 0 {
		// If no env is set, return the default
		logType = LogTypeDefault
	}
	return
}

// Get the log level from env variables
func getLogLevel() (loglevel string) {
	loglevel = os.Getenv(LogLevelEnvVarName)
	if len(loglevel) == 0 {
		// If no env is set, return the default
		loglevel = LogLevelDefault
	}
	return
}
