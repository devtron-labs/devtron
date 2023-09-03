package util

import (
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	// Logger is the defaut logger
	logger *zap.SugaredLogger
	//FIXME: remove this
	//defer Logger.Sync()
)

// Deprecated: instead calling this method inject logger from wire
func GetLogger() *zap.SugaredLogger {
	return logger
}

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic("failed to create the default logger: " + err.Error())
	}
	logger = l.Sugar()
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Generates random string
func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}
