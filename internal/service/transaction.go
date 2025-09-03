package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yusuf4ktas/backend-project/internal/domain"
	"github.com/yusuf4ktas/backend-project/internal/repository"
)

type transactionService struct {
	db              *sql.DB
	rdb             *redis.Client
	transactionRepo domain.TransactionRepository
	balanceRepo     domain.BalanceRepository
	auditService    AuditLogService
}

func NewTransactionService(db *sql.DB, rdb *redis.Client, txRepo domain.TransactionRepository, balanceRepo domain.BalanceRepository, auditService AuditLogService) TransactionService {
	return &transactionService{
		db:              db,
		rdb:             rdb,
		transactionRepo: txRepo,
		balanceRepo:     balanceRepo,
		auditService:    auditService,
	}
}

func (s *transactionService) Transfer(ctx context.Context, fromUserID int64, toUserID int64, amount float64) (*domain.Transaction, error) {
	if fromUserID == toUserID {
		return nil, fmt.Errorf("sender and receiver cannot be the same user")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("transfer amount must be a positive number")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if anything goes wrong

	// Repositories using the transaction
	balanceRepoTx := repository.NewBalanceRepository(tx, s.rdb)
	transactionRepoTx := repository.NewTransactionRepository(tx, s.rdb)

	// Debit Sender
	fromBalance, err := balanceRepoTx.GetByUserID(ctx, fromUserID)
	if err != nil {
		return nil, fmt.Errorf("could not get sender's balance: %w", err)
	}
	if fromBalance.Amount < amount {
		return nil, errors.New("insufficient funds")
	}
	fromBalance.Subtract(amount)
	fromBalance.LastUpdatedAt = time.Now()

	if err := balanceRepoTx.Update(ctx, fromBalance); err != nil {
		return nil, fmt.Errorf("failed to update sender's balance: %w", err)
	}

	// Credit Receiver
	toBalance, err := balanceRepoTx.GetByUserID(ctx, toUserID)
	if err != nil {
		return nil, fmt.Errorf("could not get receiver's balance: %w", err)
	}
	toBalance.Add(amount)
	toBalance.LastUpdatedAt = time.Now()

	if err := balanceRepoTx.Update(ctx, toBalance); err != nil {
		return nil, fmt.Errorf("failed to update receiver's balance: %w", err)
	}

	// Creation of Transaction Record
	transaction := &domain.Transaction{
		FromUserID:      fromUserID,
		ToUserID:        toUserID,
		Amount:          amount,
		TransactionType: "transfer",
		Status:          domain.StatusPending,
		CreatedAt:       time.Now(),
	}
	if err := transaction.Complete(); err != nil {
		return nil, err
	}

	if err := transactionRepoTx.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log the action
	details := fmt.Sprintf("User %d transferred %.2f to user %d", transaction.FromUserID, transaction.Amount, transaction.ToUserID)
	_, _ = s.auditService.Log(ctx, "transaction", transaction.ID, "transfer", details)

	return transaction, nil
}

func (s *transactionService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("credit amount must be a positive number")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	balanceRepoTx := repository.NewBalanceRepository(tx, s.rdb)
	transactionRepoTx := repository.NewTransactionRepository(tx, s.rdb)

	balance, err := balanceRepoTx.GetByUserID(ctx, userID)
	if err != nil {
		// User has no balance yet, create one
		if errors.Is(err, sql.ErrNoRows) {
			balance = &domain.Balance{
				UserID:        userID,
				Amount:        amount,
				LastUpdatedAt: time.Now(),
			}
			if err := balanceRepoTx.Create(ctx, balance); err != nil {
				return nil, fmt.Errorf("failed to create new balance: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get balance: %w", err)
		}
	} else {
		balance.Add(amount)
		balance.LastUpdatedAt = time.Now()
		if err := balanceRepoTx.Update(ctx, balance); err != nil {
			return nil, fmt.Errorf("failed to update balance: %w", err)
		}
	}

	//We can use 0 in receipt since SQL auto increment for db starts from 1, credit giving by admin/bank
	transaction := &domain.Transaction{
		FromUserID:      0,
		ToUserID:        userID,
		Amount:          amount,
		TransactionType: "credit",
		Status:          domain.StatusPending,
		CreatedAt:       time.Now(),
	}
	if err := transaction.Complete(); err != nil {
		return nil, err
	}

	if err := transactionRepoTx.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	details := fmt.Sprintf("User %d credited with %.2f from the bank", transaction.ToUserID, transaction.Amount)
	_, _ = s.auditService.Log(ctx, "transaction", transaction.ID, "credit", details)

	return transaction, nil
}
func (s *transactionService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("debit amount must be a positive number")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	balanceRepoTx := repository.NewBalanceRepository(tx, s.rdb)
	transactionRepoTx := repository.NewTransactionRepository(tx, s.rdb)

	balance, err := balanceRepoTx.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance.Amount < amount {
		return nil, fmt.Errorf("insufficient funds")
	}
	balance.Subtract(amount)
	balance.LastUpdatedAt = time.Now()
	if err := balanceRepoTx.Update(ctx, balance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}
	transaction := &domain.Transaction{
		FromUserID:      userID,
		ToUserID:        0,
		Amount:          amount,
		TransactionType: "debit",
		Status:          domain.StatusPending,
		CreatedAt:       time.Now(),
	}
	if err := transaction.Complete(); err != nil {
		return nil, err
	}

	if err := transactionRepoTx.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	details := fmt.Sprintf("User %d debited with %.2f to the bank", transaction.FromUserID, transaction.Amount)
	_, _ = s.auditService.Log(ctx, "transaction", transaction.ID, "debit", details)

	return transaction, nil
}

func (s *transactionService) GetTransactionHistory(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	return s.transactionRepo.GetByUserID(ctx, userID)
}

func (s *transactionService) GetByTransactionID(ctx context.Context, transactionID int64) (*domain.Transaction, error) {
	return s.transactionRepo.GetByTransactionID(ctx, transactionID)
}
