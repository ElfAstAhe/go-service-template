package postgres

const (
	// sqlCreateTableDataAuditOutbox определяет DDL-схему для создания транзакционной таблицы журнала Outbox.
	// Использует оптимизированный тип jsonb для хранения payload и типы timestamptz для учета временных зон.
	sqlCreateTableDataAuditOutbox string = `
create table if not exists data_audit_outbox (
    id varchar(50) not null,
    instance_id varchar(50) not null,
    content_type varchar(50) not null,
    payload bytea not null,
    status varchar(50) not null default 'PENDING',
    created_at timestamptz not null default now(),
    updated_at timestamptz null,
    -- 
    constraint data_audit_outbox_pk primary key(id)
)
`
	// sqlDropTableDataAuditOutbox выполняет безопасное удаление таблицы журнала Outbox из схемы данных.
	sqlDropTableDataAuditOutbox string = `
drop table if exists data_audit_outbox cascade
`
	// sqlCreateIndexDataAuditOutboxPending конструирует легковесный частичный индекс по неотправленным логам.
	// Полностью исключает сканирование диска (Seq Scan), сводя работу оператора SKIP LOCKED к наносекундам.
	sqlCreateIndexDataAuditOutboxPending string = `
create index if not exists idx_data_audit_outbox_pending on data_audit_outbox(status asc, created_at asc) where status = 'PENDING'
`
	// sqlDropIndexDataAuditOutboxPending выполняет безопасный демонтаж частичного индекса из схемы данных.
	sqlDropIndexDataAuditOutboxPending string = `
drop index if exists idx_data_audit_outbox_pending
`
	sqlCreateIndexDataAuditOutboxProcessed string = `
create index if not exists idx_data_audit_outbox_processed on data_audit_outbox(status asc, updated_at asc) where status in ('SUCCESS', 'FAILED')
`
	sqlDropIndexDataAuditOutboxProcessed string = `
drop index if exists idx_data_audit_outbox_processed
`
)
