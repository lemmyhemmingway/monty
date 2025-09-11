package models

import "time"

type Endpoint struct {
	ID        string        `json:"id"`
	URL       string        `json:"url"`
	Interval  time.Duration `json:"interval"`
	CreatedAt time.Time     `json:"created_at"`
}
