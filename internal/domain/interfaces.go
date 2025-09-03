package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	GetAllUsers(ctx context.Context) ([]User, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByUserID(ctx context.Context, userID int64) ([]Transaction, error)
	GetByTransactionID(ctx context.Context, id int64) (*Transaction, error)
}

type BalanceRepository interface {
	Create(ctx context.Context, balance *Balance) error
	GetByUserID(ctx context.Context, userID int64) (*Balance, error)
	Update(ctx context.Context, balance *Balance) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
}
