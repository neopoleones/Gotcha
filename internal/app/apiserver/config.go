package apiserver

import (
	"log"
	"os"
	"sync"

	"Gotcha/internal/app/logging"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ConfPath string
	once sync.Once
	configurationInstance GotchaConfiguration
)

// GotchaConfiguration is a simple container of presets that server really needs.
type GotchaConfiguration struct {
	AppName  string `toml:"app_name" env:"APP_NAME" env-default:"Gotcha app"`
	BindIP   string `toml:"bind_ip" env:"BIND_IP" env-default:"127.0.0.1"`
	BindPort int    `toml:"bind_port" env:"BIND_PORT" env-default:"8080"`

	// Just some nested settings
	LoggerConfiguration logging.LoggerConfiguration `toml:"logger_configuration"`
}

// NewConfiguration loads the configuration from toml file (or env variables). Panics on error.
func NewConfiguration() *GotchaConfiguration {
	once.Do(func() {
		// Check if file exists
		if _, err := os.Stat(ConfPath); os.IsNotExist(err){
			log.Panicf("Configuration %s not found: %v", ConfPath, err)
		}
		// Read content of configuration using cleanenv
		if err := cleanenv.ReadConfig(ConfPath, &configurationInstance); err != nil {
			log.Panicf("Incorrect configuration: %v", err)
		}
	})
	return &configurationInstance
}