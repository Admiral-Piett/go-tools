package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// InitLogging configures the global logger for the entire application
func InitLogging() {
	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Set custom JSON formatter with guaranteed field ordering
	log.SetFormatter(&OrderedJSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00", // ISO 8601 with milliseconds
	})

	// Enable reporting of calling function with full module path
	log.SetReportCaller(true)
}
