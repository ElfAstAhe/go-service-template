package usecase

import (
	"context"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

// OutboxListProcessedUseCase определяет интерфейс для получения сообщений outbox,
// которые ожидает отправки или обработки.
type OutboxListProcessedUseCase interface {
	// List возвращает список ожидающих обработки сообщений outbox с ограничением по количеству.
	// Возвращает срез элементов outbox или ошибку бизнес-логики в случае сбоя.
	List(ctx context.Context, updatedAt time.Time, limit int) ([]domain.Outbox[string], error)
}

// OutboxListProcessedInteractor реализует интерфейс OutboxListPendingUseCase
// и инкапсулирует бизнес-логику работы с ожидающими сообщениями outbox.
type OutboxListProcessedInteractor struct {
	repo domain.OutboxRepository[domain.Outbox[string], string]
}

var _ OutboxListProcessedUseCase = (*OutboxListProcessedInteractor)(nil)

// NewOutboxListProcessed создает и возвращает новый экземпляр OutboxListProcessedInteractor
// с внедренной зависимостью репозитория.
func NewOutboxListProcessed(repo domain.OutboxRepository[domain.Outbox[string], string]) *OutboxListPendingInteractor {
	return &OutboxListPendingInteractor{
		repo: repo,
	}
}

// List запрашивает ожидающие сообщения outbox из связанного репозитория.
// В случае ошибки на уровне хранилища, метод оборачивает её в формат BllError.
func (olp *OutboxListProcessedInteractor) List(ctx context.Context, updatedAt time.Time, limit int) ([]domain.Outbox[string], error) {
	res, err := olp.repo.ListProcessed(ctx, updatedAt, limit)
	if err != nil {
		return nil, errs.NewBllError("OutboxListProcessedInteractor.List", "failed to list processed outboxes", err)
	}

	return res, nil
}
