package service

import (
	"context"

	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type balanceService struct {
	balanceRepo domain.BalanceRepository
}

func NewBalanceService(repo domain.BalanceRepository) BalanceService {
	return &balanceService{balanceRepo: repo}
}

func (s *balanceService) GetCurrent(ctx context.Context, userID int64) (*domain.Balance, error) {
	return s.balanceRepo.GetByUserID(ctx, userID)
}
