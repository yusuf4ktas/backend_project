package domain

import (
	"fmt"
	"net/mail"
	"sync"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Indication for JSON package to always ignore this field
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Validate for user struct data
func (u *User) Validate(password string) error {
	if u.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return fmt.Errorf("invalid email address format: %w", err)
	}
	return nil
}

type Transaction struct {
	ID              int64             `json:"id"`
	FromUserID      int64             `json:"from_user_id"`
	ToUserID        int64             `json:"to_user_id"`
	Amount          float64           `json:"amount"`
	TransactionType string            `json:"transaction_type"`
	Status          TransactionStatus `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
}

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusFailed    TransactionStatus = "failed"
)

func (t *Transaction) Complete() error {
	// Only allows pending transactions to pass.
	if t.Status != StatusPending {
		return fmt.Errorf("cannot complete a transaction that is not pending (current status: %s)", t.Status)
	}

	t.Status = StatusCompleted

	return nil
}

type Balance struct {
	UserID        int64     `json:"user_id"`
	Amount        float64   `json:"amount"`
	LastUpdatedAt time.Time `json:"last_updated_at"`

	sync.RWMutex //For thread safe locking/unlocking operations
}

func (b *Balance) Add(amount float64) {
	b.Lock()
	defer b.Unlock() //For assurance to unlock in case of panic mode
	b.Amount += amount
}
func (b *Balance) Subtract(amount float64) {
	b.Lock()
	defer b.Unlock()
	b.Amount -= amount
}

type AuditLog struct {
	ID         int64     `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   int64     `json:"entity_id"`
	Action     string    `json:"action"`
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"created_at"`
}
