package model

import "time"

type Transaction struct {
	ID          int       `json:"-"`
	Login       string    `json:"-"`
	Order       string    `json:"order"`
	Amount      float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
