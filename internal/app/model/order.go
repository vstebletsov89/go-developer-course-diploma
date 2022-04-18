package model

import "time"

type Order struct {
	ID         int       `json:"-"`
	Login      string    `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
