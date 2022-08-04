package main

import (
	"flag"

	"Gotcha/internal/app/apiserver"
	logruslogger "Gotcha/internal/app/logging/logrus-logger"
)

func init() {
	// Get path of configuration file from cmdline
	flag.Parse()
	flag.StringVar(
		&apiserver.ConfPath, "cfg-path",
		"etc/default.toml", "Path to configuration file",
	)
}

func main() {
	cfg := apiserver.NewConfiguration()
	logger := logruslogger.GetLogger(&cfg.LoggerConfiguration)

	logger.Printf("Starting the %s", cfg.AppName)
	if err := apiserver.Start(cfg, logger); err != nil {
		logger.Panicf("failed to start the Gotcha apiserver: %v", err)
	}
}