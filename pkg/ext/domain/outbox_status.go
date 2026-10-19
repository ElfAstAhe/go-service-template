package domain

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// Блок констант, определяющих жизненный цикл и персистентные статусы доставки записей в транзакционном журнале Outbox.
const (
	OutboxStatusPending string = "PENDING" // Запись зафиксирована в БД и ожидает вычистки фоновым Use Case воркером
	OutboxStatusSuccess string = "SUCCESS" // Лог успешно сериализован, отправлен в брокер (Kafka) и подтвержден им (Ack)
	OutboxStatusFailed  string = "FAILED"  // В процессе доставки лога произошел критический сбой (превышено число попыток)
)

// outboxStatusMap выступает в роли легковесного хэш-сета (Set) для валидации и проверки вхождения строческих статусов за константное время O(1).
var outboxStatusMap = map[string]struct{}{
	OutboxStatusPending: {},
	OutboxStatusSuccess: {},
	OutboxStatusFailed:  {},
}

// ValidateOutboxStatus выполняет строгую доменную валидацию статуса outbox-записи.
// Возвращает типизированную ошибку errs.NewBllValidateError, если переданный маркер отсутствует в реестре разрешенных статусов.
func ValidateOutboxStatus(status string) error {
	if _, ok := outboxStatusMap[status]; !ok {
		return errs.NewBllValidateError("", fmt.Sprintf("unknown outbox status [%s]", status), nil)
	}

	return nil
}
