package model

import "time"

type Order struct {
	ID         int       `json:"-"`
	UserID     int64     `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
