package apiserver

import (
	"fmt"
	"net/http"

	"Gotcha/internal/app/logging"
)

// Start is a core function of api-server module. It opens a database connection
// configures some dependencies and fires the API server.
func Start(cfg *GotchaConfiguration, logger logging.GotchaLogger) error {
	// Get DB ...

	// Get sessions

	// Create server
	srv := newAPIServer(logger, cfg)

	// Then start the API server
	bindAddress := fmt.Sprintf("%s:%d", cfg.BindIP, cfg.BindPort)
	logger.Printf("Serving on %s", bindAddress)
	return http.ListenAndServe(bindAddress, srv.router)
}