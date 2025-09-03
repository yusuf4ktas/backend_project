package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type userRepository struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewUserRepository(db *sql.DB, rdb *redis.Client) domain.UserRepository {
	return &userRepository{
		db:  db,
		rdb: rdb,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, email, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?);`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Getting the newly generated ID from the result to avoid ID mismatch/collusion
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID after creating user: %w", err)
	}

	user.ID = id

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE email = ?;`

	row := r.db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	key := fmt.Sprintf("user:%d", id)

	cachedUser, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var user domain.User
		if json.Unmarshal([]byte(cachedUser), &user) == nil {
			return &user, nil
		}
	}

	var user domain.User
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE id = ?;`
	err = r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(&user)
	// 1 hour lifespan in cache
	r.rdb.Set(ctx, key, jsonData, 1*time.Hour)

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET username = ?, email = ?, role = ?, updated_at = ?  WHERE id = ?;`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return err
	}
	// After a successful update, invalidate the cache
	key := fmt.Sprintf("user:%d", user.ID)
	r.rdb.Del(ctx, key)
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?;`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// After a successful delete, invalidate the cache
	key := fmt.Sprintf("user:%d", id)
	r.rdb.Del(ctx, key)
	return err
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // close the rows iterator.

	var users []domain.User

	for rows.Next() {
		var user domain.User
		// For each row, scan the columns into a user struct.
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
