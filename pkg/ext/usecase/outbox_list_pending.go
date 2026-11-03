package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

type OutboxListPendingUseCase interface {
	List(ctx context.Context, limit int) ([]domain.Outbox[string], error)
}

type OutboxListPendingInteractor struct {
	repo domain.OutboxRepository[domain.Outbox[string], string]
}

func NewOutboxListPending(repo domain.OutboxRepository[domain.Outbox[string], string]) *OutboxListPendingInteractor {
	return &OutboxListPendingInteractor{
		repo: repo,
	}
}

func (olp *OutboxListPendingInteractor) List(ctx context.Context, limit int) ([]domain.Outbox[string], error) {
	res, err := olp.repo.ListPending(ctx, limit)
	if err != nil {
		return nil, errs.NewBllError("OutboxListPendingInteractor.List", "Failed to list pending outboxes", err)
	}

	return res, nil
}
