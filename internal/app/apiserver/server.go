package apiserver

import (
	"database/sql"
	"fmt"
	"net/http"

	"Gotcha/internal/app/logging"
	"Gotcha/internal/app/storage/postgres"
	"github.com/gorilla/sessions"
)

// Start is a core function of api-server module. It opens a database connection
// configures some dependencies and fires the API server.
func Start(cfg *GotchaConfiguration, logger logging.GotchaLogger) error {
	// Get DB ...
	db, err := OpenDB(cfg.DatabaseConfiguration.GetConnectionString())
	if err != nil {
		return err
	}

	// Get storage
	storage := postgres.NewStore(db)
	logger.Println("Initialized storage")
	defer storage.Close()

	// And cookie store
	// TODO: try out redis: https://github.com/boj/redistore
	cookieStore := sessions.NewCookieStore([]byte(cfg.SessionKey))

	// Create server
	srv := NewAPIServer(logger, cfg, storage, cookieStore)

	// Then fire it!
	bindAddress := fmt.Sprintf("%s:%d", cfg.BindIP, cfg.BindPort)
	logger.Printf("Serving on %s", bindAddress)
	return http.ListenAndServe(bindAddress, srv.Router)
}

func OpenDB(conStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
