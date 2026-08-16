package config

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv         string
	HTTPPort       string
	DatabaseURL    string
	JWTSecret      string
	JWTExpireHours int
	LogLevel       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("JWT_EXPIRE_HOURS", 24)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	return &Config{
		AppEnv: viper.GetString("APP_ENV"), HTTPPort: viper.GetString("HTTP_PORT"),
		DatabaseURL: viper.GetString("DATABASE_URL"), JWTSecret: viper.GetString("JWT_SECRET"),
		JWTExpireHours: viper.GetInt("JWT_EXPIRE_HOURS"), LogLevel: viper.GetString("LOG_LEVEL"),
	}, nil
}
