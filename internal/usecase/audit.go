package usecase

import (
	"context"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

type AuditUsecase struct {
	auditRepo     domain.AuditLogRepository
	random        domain.RandomGenerator
	retentionDays int
}

func NewAuditUsecase(
	auditRepo domain.AuditLogRepository,
	random domain.RandomGenerator,
	retentionDays int,
) *AuditUsecase {
	return &AuditUsecase{
		auditRepo:     auditRepo,
		random:        random,
		retentionDays: retentionDays,
	}
}

func (u *AuditUsecase) Log(ctx context.Context, userID, action, ipAddress, userAgent string, metadata map[string]any) error {
	log := &domain.AuditLog{
		ID:        u.random.GenerateUUID(),
		Action:    domain.AuditAction(action),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	if userID != "" {
		log.UserID = &userID
	}

	return u.auditRepo.Create(ctx, log)
}

func (u *AuditUsecase) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int, error) {
	return u.auditRepo.List(ctx, filter)
}

func (u *AuditUsecase) Cleanup(ctx context.Context) (int64, error) {
	before := time.Now().AddDate(0, 0, -u.retentionDays)
	return u.auditRepo.DeleteBefore(ctx, before)
}
