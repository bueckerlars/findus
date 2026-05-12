package repository

import (
	"context"

	"findus/backend/internal/domain"
)

type LabelRepository interface {
	ListAll(ctx context.Context) ([]domain.Label, error)
	GetByID(ctx context.Context, id string) (*domain.Label, error)
	Create(ctx context.Context, l *domain.Label) error
	Update(ctx context.Context, l *domain.Label) error
	Delete(ctx context.Context, id string) error
	ClearDefaultTemplateType(ctx context.Context, templateID string) error
	Count(ctx context.Context) (int64, error)
}
