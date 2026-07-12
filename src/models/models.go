package models

import "time"

type User struct {
	ID           int
	Username     string
	Email        string
	Phone        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Bot struct {
	ID            int
	UserID        int
	Username      string
	Blocked       bool
	SessionFile   string
	PaymentStatus string // "free", "pending", "paid"
	CreatedAt     time.Time
}
type ChatMessage struct {
	Role    string
	Content string
}
