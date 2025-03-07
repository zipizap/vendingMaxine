package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// OutputType defines the type of log output format
type OutputType string

const (
	// OutputTypeJSON outputs logs in JSON format
	OutputTypeJSON OutputType = "json"
	// OutputTypeText outputs logs in text format without colors
	OutputTypeText OutputType = "text"
	// OutputTypeTextColored outputs logs in text format with colors
	OutputTypeTextColored OutputType = "text-with-colors"
)

// Config contains logger configuration
type Config struct {
	Level      string     `mapstructure:"Level"`
	Type       OutputType `mapstructure:"Type"`
	TimeFormat string     `mapstructure:"TimeFormat"`
	Output     string     `mapstructure:"Output"` // "console", "file:/path/to/file", or "both:/path/to/file"
}

var defaultConfig = Config{
	Level:      "info",
	Type:       OutputTypeJSON,
	TimeFormat: time.RFC3339,
	Output:     "console",
}

// Init initializes the logger with the provided configuration
func Init(cfg *Config) {
	if cfg == nil {
		cfg = &defaultConfig
	}

	// Set global log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure time format
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Determine output destination(s)
	var output io.Writer

	// Parse output configuration
	outputParts := strings.SplitN(cfg.Output, ":", 2)
	outputType := strings.ToLower(outputParts[0])
	var filePath string
	if len(outputParts) > 1 {
		filePath = outputParts[1]
	}

	// Configure file output if needed
	var fileWriter io.Writer
	if outputType == "file" || outputType == "both" {
		if filePath == "" {
			log.Warn().Msg("File output specified but no file path provided, defaulting to console only")
		} else {
			// Ensure directory exists
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Error().Err(err).Str("directory", dir).Msg("Failed to create log directory")
			} else {
				// Open file for appending
				file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Error().Err(err).Str("file", filePath).Msg("Failed to open log file")
				} else {
					fileWriter = file
					// Don't close the file as it will be used for logging
				}
			}
		}
	}

	// Configure console output
	var consoleWriter io.Writer
	if outputType == "console" || outputType == "both" || fileWriter == nil {
		// Normalize the type value (case-insensitive)
		formatType := OutputType(strings.ToLower(string(cfg.Type)))

		switch formatType {
		case OutputTypeText, OutputTypeTextColored:
			consoleWriter = &zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: cfg.TimeFormat,
				NoColor:    formatType != OutputTypeTextColored,
			}
		case OutputTypeJSON:
			consoleWriter = os.Stderr
		default:
			// If an invalid type is provided, default to JSON
			log.Warn().Str("providedType", string(cfg.Type)).Msg("Invalid logger output type, defaulting to JSON")
			consoleWriter = os.Stderr
		}
	}

	// Combine outputs if necessary
	switch {
	case fileWriter != nil && consoleWriter != nil:
		output = io.MultiWriter(consoleWriter, fileWriter)
	case fileWriter != nil:
		output = fileWriter
	case consoleWriter != nil:
		output = consoleWriter
	default:
		// Fallback to stderr if something went wrong
		output = os.Stderr
	}

	// Set global logger
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

	// Log startup message
	log.Info().
		Str("level", cfg.Level).
		Str("type", string(cfg.Type)).
		Str("output", cfg.Output).
		Msg("Logger initialized")
}

// GetLogger returns the global logger
func GetLogger() *zerolog.Logger {
	return &log.Logger
}
