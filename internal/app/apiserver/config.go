package apiserver

import (
	"fmt"
	"log"
	"os"
	"sync"

	"Gotcha/internal/app/logging"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ConfPath              string
	once                  sync.Once
	configurationInstance GotchaConfiguration
)

// DatabaseConfiguration represents pgsql connection options
type DatabaseConfiguration struct {
	DBHost           string `toml:"host" env:"DB_HOST" env-default:"127.0.0.1"`
	DBPort           int    `toml:"port" env:"DB_PORT" env-default:"5432"`
	SSLMode          string `toml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
	Attempts         int    `toml:"attempts" env:"DB_ATTEMPTS" env-default:"5"`
	DBUsername       string `toml:"username" env:"DB_USERNAME" env-default:"postgres"`
	DBPassword       string `toml:"password" env:"DB_PASSWORD"`
	SelectedDatabase string `toml:"database" env:"DATABASE" env-default:"postgres"`
}

func (dbc *DatabaseConfiguration) GetConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbc.DBUsername, dbc.DBPassword, dbc.DBHost,
		dbc.DBPort, dbc.SelectedDatabase, dbc.SSLMode,
	)
}

type RedisConfiguration struct {
	RedisHost       string `toml:"host" env:"REDIS_HOST"`
	SessionLifetime int    `toml:"session_lifetime" env:"SESSION_LIFETIME"`
	RedisPort       int    `toml:"port" env:"REDIS_PORT"`
	IdleConnections int    `toml:"idle_connections" env:"IDLE_CONNECTIONS"`
}

// GotchaConfiguration is a simple container of presets that server really needs.
type GotchaConfiguration struct {
	AppName      string `toml:"app_name" env:"APP_NAME" env-default:"Gotcha app"`
	BindIP       string `toml:"bind_ip" env:"BIND_IP" env-default:"127.0.0.1"`
	BindPort     int    `toml:"bind_port" env:"BIND_PORT" env-default:"8080"`
	SessionKey   string `toml:"session_key" env:"SESSION_KEY" env-required:"true"`
	CookiesStore string `toml:"store" env:"STORE" env-default:"default"`

	// Just some nested settings
	LoggerConfiguration   logging.LoggerConfiguration `toml:"logger_configuration"`
	DatabaseConfiguration DatabaseConfiguration       `toml:"database_configuration"`
	RedisConfiguration    RedisConfiguration          `toml:"redis_configuration"`
}

// NewConfiguration loads the configuration from toml file (or env variables). Panics on error.
func NewConfiguration() *GotchaConfiguration {
	once.Do(func() {
		// Check if file exists
		if _, err := os.Stat(ConfPath); os.IsNotExist(err) {
			log.Panicf("Configuration %s not found: %v", ConfPath, err)
		}
		// Read content of configuration using cleanenv
		if err := cleanenv.ReadConfig(ConfPath, &configurationInstance); err != nil {
			log.Panicf("Incorrect configuration: %v", err)
		}
	})
	return &configurationInstance
}
