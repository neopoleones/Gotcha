package apiserver

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"Gotcha/internal/app/logging"
	internalStorage "Gotcha/internal/app/storage"
	"Gotcha/internal/app/storage/postgres"
	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
)

// Start is a core function of api-server module. It opens a database connection
// configures some dependencies and fires the API server.
func Start(ctx context.Context, cfg *GotchaConfiguration, logger logging.GotchaLogger) error {
	// Get DB ...
	var sessionsStore sessions.Store
	db, err := OpenDB(cfg.DatabaseConfiguration.GetConnectionString())
	if err != nil {
		return err
	}

	// Get storage
	storage := postgres.NewStore(db)
	logger.Println("Initialized storage")

	// And cookie store
	switch cfg.CookiesStore {
	case internalStorage.SessionsStoreRedis:
		conString := fmt.Sprintf("%s:%d", cfg.RedisConfiguration.RedisHost, cfg.RedisConfiguration.RedisPort)
		currStore, err := redistore.NewRediStore(
			cfg.RedisConfiguration.IdleConnections,
			"tcp", conString,
			"", []byte(cfg.SessionKey),
		)
		if err != nil {
			logger.Panicln("Failed to open redis connection.")
		}
		currStore.SetMaxAge(cfg.RedisConfiguration.SessionLifetime)
		defer func() {
			logger.Println("Closing redis store")
			_ = currStore.Close()
		}()
		sessionsStore = currStore
		logger.Printf("Connected to redis: %s", conString)
	default:
		sessionsStore = sessions.NewCookieStore([]byte(cfg.SessionKey))
	}

	// Create server
	bindAddress := fmt.Sprintf("%s:%d", cfg.BindIP, cfg.BindPort)
	srv := NewAPIServer(logger, cfg, storage, sessionsStore)
	httpServer := http.Server{
		Addr:    bindAddress,
		Handler: srv.Router,
	}

	// Then fire it in second goroutine!
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Panicln(err)
		}
	}()
	logger.Printf("Ready to serve requests on %s", bindAddress)

	// close all stores on ctx.done()
	<-ctx.Done()
	logger.Println("Shutting down the server")

	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		logger.Error(err)
	}

	mainShutdown := make(chan struct{}, 1)
	go func() {
		storage.Close()
		mainShutdown <- struct{}{}
	}()

	select {
	case <-shutdownContext.Done():
		return fmt.Errorf("server shutdown: %w", ctx.Err())
	case <-mainShutdown:
		logger.Println("Bye...")
	}
	return nil
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
