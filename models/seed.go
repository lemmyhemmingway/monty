package models

import (
	"time"

	"github.com/google/uuid"
)

// Seed inserts default endpoints if none exist
func Seed() error {
	var count int64
	DB.Model(&Endpoint{}).Count(&count)
	if count > 0 {
		return nil
	}
	ep := Endpoint{
		ID:        uuid.New().String(),
		URL:       "http://localhost:3000/health",
		Interval:  10, // seconds
		CreatedAt: time.Now(),
	}
	return DB.Create(&ep).Error
}
