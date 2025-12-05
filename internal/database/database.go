package database

import (
	"fmt"

	"github.com/gedsonn/zaapi/internal/database/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Driver string
	DSN    string
}

var DB *gorm.DB

func Initialize(cfg Config) error {
	var err error

	switch cfg.Driver {
	case "postgres":
		DB, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	default:
		return fmt.Errorf("driver não suportado: %s", cfg.Driver)
	}

	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	err = DB.AutoMigrate(
		&models.Instance{},
	)
	if err != nil {
		return err
	}

	return nil
}

func Instance() *gorm.DB {
	if DB == nil {
		panic("database: database not initialized")
	}
	return DB
}
