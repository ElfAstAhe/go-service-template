package domain

import (
	"time"

	pkgdom "github.com/ElfAstAhe/go-service-template/pkg/domain"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// DataAuditOutbox представляет собой доменную модель и системную сущность транзакционного журнала Outbox.
//
// Структура полностью сопоставляется со схемой таблицы базы данных (DDL), инкапсулирует сериализованные байты
// payload аудит-логов и реализует базовый платформенный контракт доменной сущности фреймворка.
type DataAuditOutbox struct {
	ID          string
	InstanceID  string
	ContentType string
	Payload     []byte
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Проверяем соответствие структуры строгому контракту Entity фреймворка на этапе компиляции.
var _ pkgdom.Entity[string] = (*DataAuditOutbox)(nil)
var _ Outbox[string] = (*DataAuditOutbox)(nil)

// NewEmptyDataAuditOutbox возвращает пустой экземпляр сущности со стартовым статусом PENDING.
func NewEmptyDataAuditOutbox() *DataAuditOutbox {
	return &DataAuditOutbox{
		Status: OutboxStatusPending,
	}
}

// NewDataAuditOutbox — параметризованный конструктор для явного формирования записи Outbox-журнала бизнес-данных.
func NewDataAuditOutbox(id string, instanceID string, contentType string, payload []byte, status string) *DataAuditOutbox {
	return &DataAuditOutbox{
		ID:          id,
		InstanceID:  instanceID,
		ContentType: contentType,
		Payload:     payload,
		Status:      status,
	}
}

// GetID возвращает уникальный первичный ключ текущей записи Outbox-журнала.
func (d *DataAuditOutbox) GetID() string {
	return d.ID
}

// SetID выполняет мутацию и принудительную инициализацию первичного ключа записи.
func (d *DataAuditOutbox) SetID(id string) {
	d.ID = id
}

// IsExists проверяет персистентный статус сущности (зафиксирована ли она в хранилище БД).
func (d *DataAuditOutbox) IsExists() bool {
	return d.ID != ""
}

// BeforeCreate — инфраструктурный хук (интерцептор), автоматически запускаемый СУБД/ORM перед фазой вставки (INSERT).
// Выполняет генерацию и присвоение хронологически сортируемого идентификатора UUIDv7 и выравнивает временные метки.
func (d *DataAuditOutbox) BeforeCreate() error {
	if err := pkgdom.AssignUUIDv7(d); err != nil {
		return errs.NewBllError("DataAuditOutbox.BeforeCreate", "default before create failed", err)
	}

	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	d.UpdatedAt = time.Now()

	return nil
}

// BeforeChange — infrastructure-хук (интерцептор), автоматически запускаемый СУБД/ORM перед фазой обновления (UPDATE).
func (d *DataAuditOutbox) BeforeChange() error {
	d.UpdatedAt = time.Now()

	return nil
}

// ValidateCreate выполняет доменную валидацию инвариантов и ограничений модели на этапе создания записи.
//
//goland:noinspection DuplicatedCode
func (d *DataAuditOutbox) ValidateCreate() error {
	if d.ID != "" {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateCreate", "id must be empty", nil)
	}
	if d.InstanceID == "" {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateCreate", "instance_id required", nil)
	}
	if d.ContentType == "" {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateCreate", "content_type required", nil)
	}
	if len(d.Payload) == 0 {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateCreate", "payload required", nil)
	}
	if err := ValidateOutboxStatus(d.Status); err != nil {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateCreate", "status validate failed", err)
	}

	return nil
}

// ValidateChange выполняет доменную валидацию инвариантов и ограничений модели на этапе модификации полей.
//
//goland:noinspection DuplicatedCode
func (d *DataAuditOutbox) ValidateChange() error {
	if d.ID == "" {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateChange", "id cannot be empty", nil)
	}
	if d.InstanceID == "" {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateChange", "instance_id required", nil)
	}
	if d.ContentType == "" {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateChange", "content_type required", nil)
	}
	if len(d.Payload) == 0 {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateChange", "payload required", nil)
	}
	if err := ValidateOutboxStatus(d.Status); err != nil {
		return errs.NewBllValidateError("DataAuditOutbox.ValidateChange", "status validate failed", err)
	}

	return nil
}

// GetKey возвращает идентификатор инстанса в качестве ключа распределения или партиционирования.
func (d *DataAuditOutbox) GetKey() string {
	return d.InstanceID
}

// GetContentType возвращает MIME-тип сериализованного содержимого (Payload).
func (d *DataAuditOutbox) GetContentType() string {
	return d.ContentType
}

// GetStatus возвращает текущий жизненный статус обработки Outbox-записи.
func (d *DataAuditOutbox) GetStatus() string {
	return d.Status
}

// GetPayload возвращает сырой массив байт сериализованной полезной нагрузки.
func (d *DataAuditOutbox) GetPayload() []byte {
	return d.Payload
}
