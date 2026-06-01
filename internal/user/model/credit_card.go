package model

import "time"

type CreditCard struct {
	ID        string
	UserID    string
	Token     string
	Last4     string
	Brand     string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
