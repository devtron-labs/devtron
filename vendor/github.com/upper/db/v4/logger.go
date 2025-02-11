// Copyright (c) 2012-present The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	fmtLogSessID       = `Session ID:     %05d`
	fmtLogTxID         = `Transaction ID: %05d`
	fmtLogQuery        = `Query:          %s`
	fmtLogArgs         = `Arguments:      %#v`
	fmtLogRowsAffected = `Rows affected:  %d`
	fmtLogLastInsertID = `Last insert ID: %d`
	fmtLogError        = `Error:          %v`
	fmtLogStack        = `Stack:          %v`
	fmtLogTimeTaken    = `Time taken:     %0.5fs`
	fmtLogContext      = `Context:        %v`
)

const (
	maxFrames  = 30
	skipFrames = 3
)

var (
	reInvisibleChars = regexp.MustCompile(`[\s\r\n\t]+`)
)

// LogLevel represents a verbosity level for logs
type LogLevel int8

// Log levels
const (
	LogLevelTrace LogLevel = -1

	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelPanic
)

var logLevels = map[LogLevel]string{
	LogLevelTrace: "TRACE",
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARNING",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
	LogLevelPanic: "PANIC",
}

func (ll LogLevel) String() string {
	return logLevels[ll]
}

const (
	defaultLogLevel LogLevel = LogLevelWarn
)

var defaultLogger Logger = log.New(os.Stdout, "", log.LstdFlags)

// Logger represents a logging interface that is compatible with the standard
// "log" and with many other logging libraries.
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	Print(v ...interface{})
	Printf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}

// LoggingCollector provides different methods for collecting and classifying
// log messages.
type LoggingCollector interface {
	Enabled(LogLevel) bool

	Level() LogLevel

	SetLogger(Logger)
	SetLevel(LogLevel)

	Trace(v ...interface{})
	Tracef(format string, v ...interface{})

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}

type loggingCollector struct {
	level  LogLevel
	logger Logger
}

func (c *loggingCollector) Enabled(level LogLevel) bool {
	return level >= c.level
}

func (c *loggingCollector) SetLevel(level LogLevel) {
	c.level = level
}

func (c *loggingCollector) Level() LogLevel {
	return c.level
}

func (c *loggingCollector) Logger() Logger {
	if c.logger == nil {
		return defaultLogger
	}
	return c.logger
}

func (c *loggingCollector) SetLogger(logger Logger) {
	c.logger = logger
}

func (c *loggingCollector) logf(level LogLevel, f string, v ...interface{}) {
	if level >= LogLevelPanic {
		c.Logger().Panicf(f, v...)
	}
	if level >= LogLevelFatal {
		c.Logger().Fatalf(f, v...)
	}
	if c.Enabled(level) {
		c.Logger().Printf(f, v...)
	}
}

func (c *loggingCollector) log(level LogLevel, v ...interface{}) {
	if level >= LogLevelPanic {
		c.Logger().Panic(v...)
	}
	if level >= LogLevelFatal {
		c.Logger().Fatal(v...)
	}
	if c.Enabled(level) {
		c.Logger().Print(v...)
	}
}

func (c *loggingCollector) Debugf(format string, v ...interface{}) {
	c.logf(LogLevelDebug, format, v...)
}
func (c *loggingCollector) Debug(v ...interface{}) {
	c.log(LogLevelDebug, v...)
}

func (c *loggingCollector) Tracef(format string, v ...interface{}) {
	c.logf(LogLevelTrace, format, v...)
}
func (c *loggingCollector) Trace(v ...interface{}) {
	c.log(LogLevelDebug, v...)
}

func (c *loggingCollector) Infof(format string, v ...interface{}) {
	c.logf(LogLevelInfo, format, v...)
}
func (c *loggingCollector) Info(v ...interface{}) {
	c.log(LogLevelInfo, v...)
}

func (c *loggingCollector) Warnf(format string, v ...interface{}) {
	c.logf(LogLevelWarn, format, v...)
}
func (c *loggingCollector) Warn(v ...interface{}) {
	c.log(LogLevelWarn, v...)
}

func (c *loggingCollector) Errorf(format string, v ...interface{}) {
	c.logf(LogLevelError, format, v...)
}
func (c *loggingCollector) Error(v ...interface{}) {
	c.log(LogLevelError, v...)
}

func (c *loggingCollector) Fatalf(format string, v ...interface{}) {
	c.logf(LogLevelFatal, format, v...)
}
func (c *loggingCollector) Fatal(v ...interface{}) {
	c.log(LogLevelFatal, v...)
}

func (c *loggingCollector) Panicf(format string, v ...interface{}) {
	c.logf(LogLevelPanic, format, v...)
}
func (c *loggingCollector) Panic(v ...interface{}) {
	c.log(LogLevelPanic, v...)
}

var defaultLoggingCollector LoggingCollector = &loggingCollector{
	level:  defaultLogLevel,
	logger: defaultLogger,
}

// QueryStatus represents the status of a query after being executed.
type QueryStatus struct {
	SessID uint64
	TxID   uint64

	RowsAffected *int64
	LastInsertID *int64

	RawQuery string
	Args     []interface{}

	Err error

	Start time.Time
	End   time.Time

	Context context.Context
}

func (q *QueryStatus) Query() string {
	query := reInvisibleChars.ReplaceAllString(q.RawQuery, " ")
	query = strings.TrimSpace(query)
	return query
}

func (q *QueryStatus) Stack() []string {
	frames := collectFrames()
	lines := make([]string, 0, len(frames))

	for _, frame := range frames {
		lines = append(lines, fmt.Sprintf("%s@%s:%d", frame.Function, frame.File, frame.Line))
	}
	return lines
}

// String returns a formatted log message.
func (q *QueryStatus) String() string {
	lines := make([]string, 0, 8)

	if q.SessID > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogSessID, q.SessID))
	}

	if q.TxID > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogTxID, q.TxID))
	}

	if query := q.RawQuery; query != "" {
		lines = append(lines, fmt.Sprintf(fmtLogQuery, q.Query()))
	}

	if len(q.Args) > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogArgs, q.Args))
	}

	if stack := q.Stack(); len(stack) > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogStack, "\n\t"+strings.Join(stack, "\n\t")))
	}

	if q.RowsAffected != nil {
		lines = append(lines, fmt.Sprintf(fmtLogRowsAffected, *q.RowsAffected))
	}
	if q.LastInsertID != nil {
		lines = append(lines, fmt.Sprintf(fmtLogLastInsertID, *q.LastInsertID))
	}

	if q.Err != nil {
		lines = append(lines, fmt.Sprintf(fmtLogError, q.Err))
	}

	lines = append(lines, fmt.Sprintf(fmtLogTimeTaken, float64(q.End.UnixNano()-q.Start.UnixNano())/float64(1e9)))

	if q.Context != nil {
		lines = append(lines, fmt.Sprintf(fmtLogContext, q.Context))
	}

	return "\t" + strings.Replace(strings.Join(lines, "\n"), "\n", "\n\t", -1) + "\n\n"
}

// LC returns the logging collector.
func LC() LoggingCollector {
	return defaultLoggingCollector
}

func init() {
	if logLevel := strings.ToUpper(os.Getenv("UPPER_DB_LOG")); logLevel != "" {
		for ll := range logLevels {
			if ll.String() == logLevel {
				LC().SetLevel(ll)
				break
			}
		}
	}
}

func collectFrames() []runtime.Frame {
	pc := make([]uintptr, maxFrames)
	n := runtime.Callers(skipFrames, pc)
	if n == 0 {
		return nil
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	collectedFrames := make([]runtime.Frame, 0, maxFrames)
	discardedFrames := make([]runtime.Frame, 0, maxFrames)
	for {
		frame, more := frames.Next()

		// collect all frames except those from upper/db and runtime stack
		if (strings.Contains(frame.Function, "upper/db") || strings.Contains(frame.Function, "/go/src/")) && !strings.Contains(frame.Function, "test") {
			discardedFrames = append(discardedFrames, frame)
		} else {
			collectedFrames = append(collectedFrames, frame)
		}

		if !more {
			break
		}
	}

	if len(collectedFrames) < 1 {
		return discardedFrames
	}

	return collectedFrames
}
