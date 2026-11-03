package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

type OutboxChangeStatusUseCase interface {
	Change(ctx context.Context, id string, status string) error
}

type OutboxChangeStatusInteractor struct {
	uw   domain.UnitOfWork
	repo extdom.OutboxRepository[extdom.Outbox[string], string]
}

func NewOutboxChangeStatus(uw domain.UnitOfWork, repo extdom.OutboxRepository[extdom.Outbox[string], string]) *OutboxChangeStatusInteractor {
	return &OutboxChangeStatusInteractor{
		uw:   uw,
		repo: repo,
	}
}

func (ocs *OutboxChangeStatusInteractor) Change(ctx context.Context, id string, status string) error {
	// fail fast
	if strings.TrimSpace(status) == "" {
		return errs.NewInvalidArgumentError("id", "empty")
	}
	if err := extdom.ValidateOutboxStatus(status); err != nil {
		return errs.NewInvalidArgumentErrorChain("status", status, err)
	}
	// act
	err := ocs.uw.Execute(ctx, func(uwCtx context.Context) error {
		if txErr := ocs.repo.ChangeStatus(uwCtx, id, status); txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return errs.NewBllError("OutboxChangeStatusInteractor.Change", fmt.Sprintf("change outbox id [%s] status [%s] failed", id, status), err)
	}

	return nil
}
