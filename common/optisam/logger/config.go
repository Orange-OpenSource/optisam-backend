package logger

// Config holds details necessary for logging.
type Config struct {

	// Level is the minimum log level that should appear on the output.
	LogLevel int

	// LogTimeFormat Time Format.
	LogTimeFormat string
}
