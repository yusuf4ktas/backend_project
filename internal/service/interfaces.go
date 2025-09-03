package service

import (
	"context"

	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type UserService interface {
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	GetByID(ctx context.Context, userID int64) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	Delete(ctx context.Context, userID int64) error
}
type TransactionService interface {
	Transfer(ctx context.Context, fromUserID int64, toUserID int64, amount float64) (*domain.Transaction, error)
	Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	GetTransactionHistory(ctx context.Context, userID int64) ([]domain.Transaction, error)
	GetByTransactionID(ctx context.Context, id int64) (*domain.Transaction, error)
}

type BalanceService interface {
	GetCurrent(ctx context.Context, userID int64) (*domain.Balance, error)
}

type AuditLogService interface {
	Log(ctx context.Context, entityType string, entityID int64, action string, details string) (*domain.AuditLog, error)
}
