package models

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	if dsn == "" {
		// Use SQLite for local development
		DB, err = gorm.Open(sqlite.Open("monty.db"), &gorm.Config{})
	} else {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := DB.AutoMigrate(&Endpoint{}, &Status{}, &SSLStatus{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}
}
