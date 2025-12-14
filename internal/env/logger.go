package env

import (
	"log"
	"os"
)

const (
	LoggerFlags = log.LstdFlags | log.Lshortfile
)

func init() {
	log.Default().SetFlags(LoggerFlags)
}

func NewLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, LoggerFlags)
}
