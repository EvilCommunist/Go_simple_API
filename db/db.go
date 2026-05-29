package db

import (
	"fmt"

	"github.com/goloop/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string `env:"DB_HOST" def:"NONE"`
	Login    string `env:"DB_USER" def:"NONE"`
	Password string `env:"DB_PASSWORD" def:"NONE"`
	Database string `env:"DB_NAME" def:"NONE"`
	Port     string `env:"DB_PORT" def:"NONE"`
}

func Connect() (*gorm.DB, error) {
	if err := env.Load(".env"); err != nil {
		return nil, fmt.Errorf("Unable to get DB parameters from .env")
	}

	var cfg Config
	if err := env.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("Unable to parse DB parameters")
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.Login, cfg.Password, cfg.Database, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
