package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

// OutboxListPendingUseCase определяет интерфейс для получения сообщений outbox,
// которые ожидает отправки или обработки.
type OutboxListPendingUseCase interface {
	// List возвращает список ожидающих обработки сообщений outbox с ограничением по количеству.
	// Возвращает срез элементов outbox или ошибку бизнес-логики в случае сбоя.
	List(ctx context.Context, limit int) ([]domain.Outbox[string], error)
}

// OutboxListPendingInteractor реализует интерфейс OutboxListPendingUseCase
// и инкапсулирует бизнес-логику работы с ожидающими сообщениями outbox.
type OutboxListPendingInteractor struct {
	repo domain.OutboxRepository[domain.Outbox[string], string]
}

// NewOutboxListPending создает и возвращает новый экземпляр OutboxListPendingInteractor
// с внедренной зависимостью репозитория.
func NewOutboxListPending(repo domain.OutboxRepository[domain.Outbox[string], string]) *OutboxListPendingInteractor {
	return &OutboxListPendingInteractor{
		repo: repo,
	}
}

// List запрашивает ожидающие сообщения outbox из связанного репозитория.
// В случае ошибки на уровне хранилища, метод оборачивает её в формат BllError.
func (olp *OutboxListPendingInteractor) List(ctx context.Context, limit int) ([]domain.Outbox[string], error) {
	res, err := olp.repo.ListPending(ctx, limit)
	if err != nil {
		return nil, errs.NewBllError("OutboxListPendingInteractor.List", "failed to list pending outboxes", err)
	}

	return res, nil
}
