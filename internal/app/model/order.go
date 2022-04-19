package model

import "time"

type Order struct {
	ID         int       `json:"-"`
	Login      string    `json:"-"`
	Number     string    `json:"order"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
