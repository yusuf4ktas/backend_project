package repository

import (
	"context"
	"database/sql"

	"github.com/yusuf4ktas/backend-project/internal/domain"
)

type auditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `INSERT INTO audit_logs (entity_type, entity_id, action, details, created_at) VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		log.EntityType,
		log.EntityID,
		log.Action,
		log.Details,
		log.CreatedAt,
	)
	return err
}
