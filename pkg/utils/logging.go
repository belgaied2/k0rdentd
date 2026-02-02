package utils

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// PlainFormatter is a custom formatter that only returns the message
type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Return only the message followed by a newline
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}

// SetupLogging configures logging based on debug flag
func SetupLogging(debug bool) {
	log = logrus.New()

	if debug {
		log.SetLevel(logrus.DebugLevel)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
		log.SetOutput(os.Stdout)
	} else {
		log.SetLevel(logrus.InfoLevel)
		log.SetFormatter(new(PlainFormatter))
		log.SetOutput(os.Stderr)
	}
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	return log
}
