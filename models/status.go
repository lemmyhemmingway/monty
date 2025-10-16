package models

import "time"

type Status struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	EndpointID    string    `json:"endpoint_id"`
	Code          int       `json:"code"`
	ResponseTime  int       `json:"response_time"`  // milliseconds
	ErrorMessage  string    `json:"error_message"`
	CheckedAt     time.Time `json:"checked_at"`
}
