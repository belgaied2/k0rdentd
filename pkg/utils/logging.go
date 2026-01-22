package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

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
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		log.SetOutput(os.Stderr)
	}
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	return log
}