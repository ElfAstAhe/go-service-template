package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

// OutboxChangeStatusUseCase определяет интерфейс для изменения статуса
// сообщений в паттерне Outbox.
type OutboxChangeStatusUseCase interface {
	// Change изменяет статус сообщения по его идентификатору.
	// Возвращает ошибку валидации или ошибку бизнес-логики в случае сбоя.
	Change(ctx context.Context, id string, status string) error
}

// OutboxChangeStatusInteractor реализует интерфейс OutboxChangeStatusUseCase,
// обеспечивая атомарное изменение статуса в рамках транзакции Unit of Work.
type OutboxChangeStatusInteractor struct {
	uw   domain.UnitOfWork
	repo extdom.OutboxRepository[extdom.Outbox[string], string]
}

var _ OutboxChangeStatusUseCase = (*OutboxChangeStatusInteractor)(nil)

// NewOutboxChangeStatus создает и возвращает новый экземпляр OutboxChangeStatusInteractor
// с необходимыми зависимостями для управления транзакциями и репозиторием.
func NewOutboxChangeStatus(uw domain.UnitOfWork, repo extdom.OutboxRepository[extdom.Outbox[string], string]) *OutboxChangeStatusInteractor {
	return &OutboxChangeStatusInteractor{
		uw:   uw,
		repo: repo,
	}
}

// Change выполняет валидацию входных данных и обновляет статус сообщения.
// Операция обновления репозитория оборачивается в транзакцию Unit of Work.
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
