package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"Gotcha/internal/app/apiserver"
	logruslogger "Gotcha/internal/app/logging/logrus-logger"
	"github.com/asaskevich/govalidator"
)

func init() {
	// Get path of configuration file from cmdline
	flag.Parse()
	flag.StringVar(
		&apiserver.ConfPath, "cfg-path",
		"etc/default.toml", "Path to configuration file",
	)
	govalidator.SetFieldsRequiredByDefault(true)
}

func main() {
	// We wanna graceful shutdown, right?
	interruptContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := apiserver.NewConfiguration()
	logger := logruslogger.GetLogger(&cfg.LoggerConfiguration)

	logger.Printf("Starting the %s", cfg.AppName)
	if err := apiserver.Start(interruptContext, cfg, logger); err != nil {
		logger.Panicf("failed to start the Gotcha apiserver: %v", err)
	}
}
