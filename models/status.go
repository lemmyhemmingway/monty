package models

import "time"

type Status struct {
	ID         string    `json:"id"`
	EndpointID string    `json:"endpoint_id"`
	Code       int       `json:"code"`
	CheckedAt  time.Time `json:"checked_at"`
}
