package service

import (
	"context"
	"time"

	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type auditLogService struct {
	auditLogRepo domain.AuditLogRepository
}

func NewAuditLogService(repo domain.AuditLogRepository) AuditLogService {
	return &auditLogService{
		auditLogRepo: repo,
	}
}

func (s *auditLogService) Log(ctx context.Context, entityType string, entityID int64, action string, details string) (*domain.AuditLog, error) {
	log := &domain.AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Details:    details,
		CreatedAt:  time.Now(),
	}
	err := s.auditLogRepo.Create(ctx, log)
	if err != nil {
		return nil, err
	}
	return log, nil
}
