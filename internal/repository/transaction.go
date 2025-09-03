package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type transactionRepository struct {
	db  DBTX
	rdb *redis.Client
}

func NewTransactionRepository(db DBTX, rdb *redis.Client) *transactionRepository {
	return &transactionRepository{
		db:  db,
		rdb: rdb,
	}
}

func (tr *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, status, created_at) VALUES (?,?,?,?,?,?);`

	_, err := tr.db.ExecContext(
		ctx,
		query,
		tx.FromUserID,
		tx.ToUserID,
		tx.Amount,
		tx.TransactionType,
		tx.Status,
		time.Now(),
	)
	return err
}

func (tr *transactionRepository) GetByUserID(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, transaction_type, status, created_at FROM transactions WHERE from_user_id = ? OR to_user_id = ?;`

	rows, err := tr.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.FromUserID,
			&tx.ToUserID,
			&tx.Amount,
			&tx.TransactionType,
			&tx.Status,
			&tx.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (tr *transactionRepository) GetByTransactionID(ctx context.Context, transactionID int64) (*domain.Transaction, error) {
	key := fmt.Sprintf("transaction:%d", transactionID)

	// Try to get the transaction from the Redis cache
	cachedTx, err := tr.rdb.Get(ctx, key).Result()
	if err == nil {
		//Cache found case, unmarshal and return the cached data
		var transaction domain.Transaction
		if json.Unmarshal([]byte(cachedTx), &transaction) == nil {
			return &transaction, nil
		}
	}

	// Cache not found, get transaction from database
	var transaction domain.Transaction
	query := `SELECT id, from_user_id, to_user_id, amount, transaction_type, status, created_at FROM transactions WHERE id = ?;`
	err = tr.db.QueryRowContext(ctx, query, transactionID).Scan(
		&transaction.ID,
		&transaction.FromUserID,
		&transaction.ToUserID,
		&transaction.Amount,
		&transaction.TransactionType,
		&transaction.Status,
		&transaction.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Store the new data in the cache
	jsonData, _ := json.Marshal(&transaction)
	// 24 hours lifespan in cache.
	tr.rdb.Set(ctx, key, jsonData, 24*time.Hour)

	return &transaction, nil
}
