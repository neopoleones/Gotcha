package logging

import (
	"github.com/sirupsen/logrus"
)

// GotchaLogger ...
// In fact, only logrus-like loggers can be used here, because I wanna Fields* support.
type GotchaLogger interface {
	logrus.StdLogger
	logrus.FieldLogger
}

// LoggerConfiguration represents most valuable options of logger
type LoggerConfiguration struct {
	Level      string   `toml:"level" env:"LOGGER_LEVEL" env-default:"info"`
	ShowCaller bool     `toml:"show_caller" env:"SHOW_CALLER" env-default:"true"`
	OutputFormat string `toml:"output_format" env:"LOGGER_OUTPUT_FORMAT"`
	// ... other field will be added later
}

const (
	OutputFormatColored  = "terminal-colored"
	OutputFormatTerminal = "terminal"
	OutputFormatJson     = "json"
)