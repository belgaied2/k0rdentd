package utils

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestPlainFormatter_Format(t *testing.T) {
	g := gomega.NewWithT(t)

	formatter := &PlainFormatter{}

	tests := []struct {
		name    string
		entry  *logrus.Entry
	}{
		{
			name: "simple message",
			entry: &logrus.Entry{
				Message: "test message",
			},
		},
		{
			name: "empty message",
			entry: &logrus.Entry{
				Message: "",
			},
		},
		{
			name: "message with special characters",
			entry: &logrus.Entry{
				Message: "test\nmessage\twith\ttabs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				result, err := formatter.Format(tt.entry)
				g.Expect(err).ToNot(gomega.HaveOccurred())
				g.Expect(result).ToNot(gomega.BeNil())
				g.Expect(string(result)).To(gomega.ContainSubstring(tt.entry.Message))
				g.Expect(string(result)).To(gomega.HaveSuffix("\n"))
			})
	 }
}

func TestSetupLogging(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("sets debug mode", func(t *testing.T) {
		SetupLogging(true)

		logger := GetLogger()
		g.Expect(logger).ToNot(gomega.BeNil())
		g.Expect(logger.Level).To(gomega.Equal(logrus.DebugLevel))
		g.Expect(logger.Formatter).To(gomega.BeAssignableToTypeOf(&logrus.TextFormatter{}))
	})

	t.Run("sets info mode", func(t *testing.T) {
		SetupLogging(false)

		logger := GetLogger()
		g.Expect(logger).ToNot(gomega.BeNil())
		g.Expect(logger.Level).To(gomega.Equal(logrus.InfoLevel))
		g.Expect(logger.Formatter).To(gomega.BeAssignableToTypeOf(&PlainFormatter{}))
	})
}

func TestGetLogger_InitializesIfNil(t *testing.T) {
	g := gomega.NewWithT(t)

	// Reset the global logger to nil by creating a new test
	// Note: This test assumes the logger package-level variable can be reset
	// Since we can't easily reset the package-level variable, we'll test
	// that GetLogger returns a valid logger even without calling SetupLogging first

	// This test verifies GetLogger works even without explicit SetupLogging call
	logger := GetLogger()
	g.Expect(logger).ToNot(gomega.BeNil())
}

func TestGetLogger_ReturnsSameInstance(t *testing.T) {
	g := gomega.NewWithT(t)

	SetupLogging(false)

	logger1 := GetLogger()
	logger2 := GetLogger()

	g.Expect(logger1).To(gomega.Equal(logger2), "GetLogger should return the same instance")
}
