package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type balanceRepository struct {
	db  DBTX
	rdb *redis.Client
}

func NewBalanceRepository(db DBTX, rdb *redis.Client) domain.BalanceRepository {
	return &balanceRepository{
		db:  db,
		rdb: rdb,
	}
}

func (r *balanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	query := `INSERT INTO balances (user_id, amount, last_updated_at) VALUES (?, ?, ?);`

	_, err := r.db.ExecContext(
		ctx,
		query,
		balance.UserID,
		balance.Amount,
		balance.LastUpdatedAt,
	)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("balance:user:%d", balance.UserID)
	r.rdb.Del(ctx, key)

	return nil
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	key := fmt.Sprintf("balance:user:%d", userID)

	// Try to get the balance from the Redis cache
	cachedBalance, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		//Cache found case, unmarshal and return the cached data.
		var balance domain.Balance
		if json.Unmarshal([]byte(cachedBalance), &balance) == nil {
			return &balance, nil
		}
	}

	// Cache not found, get balance from database.
	var balance domain.Balance
	query := `SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = ?;`
	err = r.db.QueryRowContext(ctx, query, userID).Scan(&balance.UserID, &balance.Amount, &balance.LastUpdatedAt)
	if err != nil {
		return nil, err
	}

	// Store the new data in the cache.
	jsonData, _ := json.Marshal(&balance)
	// 15 minute lifespan in cache.
	r.rdb.Set(ctx, key, jsonData, 15*time.Minute)

	return &balance, nil
}

func (r *balanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	query := `UPDATE balances SET amount = ?, last_updated_at = ? WHERE user_id = ?;`

	_, err := r.db.ExecContext(
		ctx,
		query,
		balance.Amount,
		balance.LastUpdatedAt,
		balance.UserID,
	)
	if err != nil {
		return err
	}

	// After a successful write, invalidating the cache.
	key := fmt.Sprintf("balance:user:%d", balance.UserID)
	r.rdb.Del(ctx, key)

	return nil
}
