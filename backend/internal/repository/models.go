package repository

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type RefreshTokenRecord struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type RatePoint struct {
	Date string  `json:"date"`
	Rate float64 `json:"rate"`
}

type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
