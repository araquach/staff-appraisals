package util

import (
	"log"
	"os"
)

// NewLogger creates a standard logger with timestamps.
func NewLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}
