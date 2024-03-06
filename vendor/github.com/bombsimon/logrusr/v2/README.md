# Logrusr

[![Coverage Status](https://coveralls.io/repos/github/bombsimon/logrusr/badge.svg?branch=master)](https://coveralls.io/github/bombsimon/logrusr?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/bombsimon/logrusr)](https://goreportcard.com/report/github.com/bombsimon/logrusr)
[![golangci-lint](https://golangci.com/badges/github.com/bombsimon/logrusr.svg)](https://golangci.com/r/github.com/bombsimon/logrusr)

A [logr](https://github.com/go-logr/logr) implementation using
[logrus](https://github.com/sirupsen/logrus).

## Usage

```go
import (
    "github.com/bombsimon/logrusr/v2"
    "github.com/go-logr/logr"
    "github.com/sirupsen/logrus"
)

func main() {
    logrusLog := logrus.New()
    log := logrusr.NewLogger(logrusLog)

    log = log.WithName("MyName").WithValues("user", "you")
    log.Info("Logr in action!", "the answer", 42)
}
```

For more details, see [example](example/main.go).

## Implementation details

The NewLogger method takes a `logrus.FieldLogger` interface as input which means
this works with both `logrus.Logger` and `logrus.Entry`. This is currently a
quite naive implementation in early state. Use with caution.
