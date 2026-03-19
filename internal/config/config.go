package config

import (
	"github.com/spf13/viper"
)

type Config struct {
    DBHost   string
    DBName   string
    DBUser   string
    DBPass   string
    SSL      string
    ChanBind string
    BaseURL  string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

    config := &Config{
        DBHost:   viper.GetString("DATABASE_HOST"),
        DBName:   viper.GetString("DATABASE_NAME"),
        DBUser:   viper.GetString("DATABASE_USER"),
        DBPass:   viper.GetString("DATABASE_PASSWORD"),
        SSL:      viper.GetString("SSL"),
        ChanBind: viper.GetString("CHANNEL_BINDING"),
        BaseURL:  viper.GetString("BASE_URL"),
    }

    return config, nil
}
