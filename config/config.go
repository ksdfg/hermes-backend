package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	// Whatsapp client version
	VersionMajor int `mapstructure:"WHATSAPP_VERSION_MAJOR"`
	VersionMinor int `mapstructure:"WHATSAPP_VERSION_MINOR"`
	VersionPatch int `mapstructure:"WHATSAPP_VERSION_PATCH"`

	// App client info
	ClientLong    string `mapstructure:"CLIENT_LONG"`
	ClientShort   string `mapstructure:"CLIENT_SHORT"`
	ClientVersion string `mapstructure:"CLIENT_VERSION"`

	// Size of QR code generated
	QrSize int `mapstructure:"QR_SIZE"`

	// Peak Concurrency
	Concurrency int `mapstructure:"CONCURRENCY"`

	// Origins Allowed by CORS
	AllowOrigins string `mapstructure:"ALLOW_ORIGINS"`
}

// Unexported variable to implement singleton pattern
var cfg *Config = nil

/*
Init will read all config variables from the .env and environment variables
*/
func Init() {
	// Setup viper
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Set default values for config vars
	viper.SetDefault("WHATSAPP_VERSION_MAJOR", -1)
	viper.SetDefault("WHATSAPP_VERSION_MINOR", -1)
	viper.SetDefault("WHATSAPP_VERSION_MINOR", -1)
	viper.SetDefault("CLIENT_LONG", "")
	viper.SetDefault("CLIENT_SHORT", "")
	viper.SetDefault("CLIENT_VERSION", "")
	viper.SetDefault("QR_SIZE", -1)
	viper.SetDefault("CONCURRENCY", -1)
	viper.SetDefault("ALLOW_ORIGINS", "")

	// Automatically override values in config file with those in environment
	viper.AutomaticEnv()

	// Read config file
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			log.Fatal(err)
		}
	}

	// Set config object
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Make sure all the config variables are set
	if cfg.VersionMajor == -1 || cfg.VersionMinor == -1 || cfg.VersionPatch == -1 {
		log.Fatal("whatsapp version configuration not set in env")
	}
	if cfg.ClientLong == "" || cfg.ClientShort == "" || cfg.ClientVersion == "" {
		log.Fatal("app client configuration not set in env")
	}
	if cfg.QrSize == -1 {
		log.Fatal("qr size not set in env")
	}
	if cfg.Concurrency == -1 {
		log.Fatal("peak concurrency not set in env")
	}
	if cfg.AllowOrigins == "" {
		log.Fatal("origins allowed by CORS not set in env")
	}
}

/*
Get will return the config object set in Init
*/
func Get() *Config {
	return cfg
}
