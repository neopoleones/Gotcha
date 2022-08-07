package logrus_logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"Gotcha/internal/app/logging"
	"github.com/sirupsen/logrus"
)

var (
	logrusInstance *logrus.Logger
	once           sync.Once
)

// enrichCallerInformation is a custom caller prettyfier. It takes current frame and returns
// 	function in format "[module.function:line]", file in format "directory/filename.go".
func enrichCallerInformation(frame *runtime.Frame) (function string, file string) {
	dir, fileName := path.Split(frame.File)
	trimmedDir := path.Base(dir)

	funcName := fmt.Sprintf("[%s:%d]", path.Base(frame.Function), frame.Line)
	return funcName, path.Join(trimmedDir, fileName)
}

// GetLogger initializes logrus instance and applies some presets depending on logger configuration.
func GetLogger(cfg *logging.LoggerConfiguration) *logrus.Entry {
	once.Do(func() {
		logrusInstance = logrus.New()
		// Setup logrus and apply configuration options
		logrusInstance.SetReportCaller(cfg.ShowCaller)

		level, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			log.Panicf("GetLogger failure: %v", err)
		}

		logrusInstance.SetLevel(level)
		logrusInstance.SetOutput(os.Stdout)

		// Set output format
		if cfg.OutputFormat == logging.OutputFormatJson {
			logrusInstance.SetFormatter(&logrus.JSONFormatter{
				CallerPrettyfier: enrichCallerInformation,
			})
		} else {
			customTextFormatter := &logrus.TextFormatter{
				FullTimestamp: true,
				ForceQuote:    true,
			}

			if cfg.OutputFormat == logging.OutputFormatTerminal {
				customTextFormatter.DisableColors = true
			} else if cfg.OutputFormat == logging.OutputFormatColored {
				customTextFormatter.CallerPrettyfier = enrichCallerInformation
			}

			logrusInstance.SetFormatter(customTextFormatter)
		}
	})
	return logrus.NewEntry(logrusInstance)
}
