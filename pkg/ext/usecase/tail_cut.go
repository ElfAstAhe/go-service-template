package usecase

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	extdom "github.com/ElfAstAhe/go-service-template/pkg/ext/domain"
)

// TailCutUseCase определяет строго типизированный обобщенный интерфейс прикладного уровня (Use Case)
// для координации и выполнения операций попоточного удаления устаревших записей из журнала Outbox.
//
// Выступает в роли оркестратора бизнес-логики паттерна Tail Cutter, изолируя фоновые инфраструктурные
// воркеры (транспортный слой) от прямых вызовов к базам данных и слою репозиториев.
type TailCutUseCase[ID comparable] interface {
	// Cut осуществляет прикладную операцию удаления одной конкретной записи Outbox по её идентификатору.
	// Метод вызывается асинхронным воркером для попоточного удаления, инкапсулируя внутри себя логику логирования и сбора телеметрии.
	Cut(ctx context.Context, id ID) error
}

// TailCutInteractor является конкретной реализацией интерфейса TailCutUseCase.
//
// Оркестрирует фазу физической очистки устаревших записей, инжектируя механизм UnitOfWork
// для выполнения каждой операции удаления в рамках изолированной и безопасной ACID-транзакции СУБД.
type TailCutInteractor[ID comparable] struct {
	uw       domain.UnitOfWork
	tailRepo extdom.TailRepository[ID]
}

// Проверяем соответствие интерактора контракту UseCase на этапе компиляции для строковых ключей
var _ TailCutUseCase[string] = (*TailCutInteractor[string])(nil)

// NewTailCutUseCase — фабричный конструктор прикладного интерактора контура очистки хвостов.
func NewTailCutUseCase[ID comparable](
	uw domain.UnitOfWork,
	tailRepo extdom.TailRepository[ID],
) *TailCutInteractor[ID] {
	return &TailCutInteractor[ID]{
		uw:       uw,
		tailRepo: tailRepo,
	}
}

// Cut выполняет транзакционное удаление одной записи Outbox по её идентификатору, оборачивая операцию в UnitOfWork.
func (tc *TailCutInteractor[ID]) Cut(ctx context.Context, id ID) error {
	err := tc.uw.Execute(ctx, func(txCtx context.Context) error {
		txErr := tc.tailRepo.Delete(txCtx, id)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return errs.NewBllError("TailCutInteractor.Cut", "cut tail", err)
	}

	return nil
}
