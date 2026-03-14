package config

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost   string
	DBName   string
	DBUser   string
	DBPass   string
	SSL      string
	ChanBind string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	var fileLookupError viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &fileLookupError) {
			return nil, fileLookupError
		} else {
			return nil, err
		}
	}

	config := &Config{
			DBHost:   viper.GetString("DATABASE_HOST"),
			DBName:   viper.GetString("DATABASE_NAME"),
			DBUser:   viper.GetString("DATABASE_USER"),
			DBPass:   viper.GetString("DATABASE_PASSWORD"),
			SSL:      viper.GetString("SSL"),
			ChanBind: viper.GetString("CHANNEL_BINDING"),
        }

        return config, nil
}
