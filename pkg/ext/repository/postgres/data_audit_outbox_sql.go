package postgres

// Блок констант содержит SQL-запросы для работы с таблицей аудита данных data_audit_outbox.
const (
	// SqlDataAuditOutboxFind возвращает SQL-запрос для поиска одной записи outbox по её идентификатору.
	SqlDataAuditOutboxFind string = `
select
    id,
    instance_id,
    content_type,
    payload,
    status,
    created_at,
    updated_at
from
    data_audit_outbox
where
    id = $1
`
	// SqlDataAuditOutboxList возвращает SQL-запрос для постраничной выборки всех записей outbox,
	// отсортированных по времени создания и идентификатору.
	SqlDataAuditOutboxList string = `
select
    id,
    instance_id,
    content_type,
    payload,
    status,
    created_at,
    updated_at
from
    data_audit_outbox
order by
    created_at asc,
    id asc
offset $2
limit $1
`
	// SqlDataAuditOutboxCreate возвращает SQL-запрос для вставки новой записи в таблицу outbox.
	SqlDataAuditOutboxCreate string = `
insert into data_audit_outbox (
    id,
    instance_id,
    content_type,
    payload,
    status,
    created_at
)
values (
    $1, $2, $3, $4, $5, $6
)
`
	// SqlDataAuditOutboxChange возвращает SQL-запрос для полного обновления всех полей существующей записи outbox.
	SqlDataAuditOutboxChange string = `
update
    data_audit_outbox
set
    instance_id = $2,
    content_type = $3,
    payload = $4,
    status = $5,
    updated_at = $6
where
    id = $1
`
	// SqlDataAuditOutboxDelete возвращает SQL-запрос для удаления записи outbox по идентификатору.
	SqlDataAuditOutboxDelete string = `
delete
from
    data_audit_outbox
where
    id = $1
`
	// SqlDataAuditOutboxGetTail возвращает SQL-запрос для получения идентификаторов устаревших записей
	// с определенными статусами для последующей очистки (в запросе опечатка в имени таблицы `data_audit_aoutbox`).
	SqlDataAuditOutboxGetTail string = `
select
    id
from
    data_audit_aoutbox
where
    status = any($1)
and updated_at < $2
limit $3
`
	// SqlDataAuditOutboxListPending возвращает SQL-запрос для получения списка необработанных записей outbox
	// с фильтрацией по статусам и сортировкой по хронологии.
	SqlDataAuditOutboxListPending string = `
select
    id,
    instance_id,
    content_type,
    payload,
    status,
    created_at,
    updated_at
from
    data_audit_outbox
where
    status = any($1)
order by
    created_at asc
limit $2
`
	// SqlDataAuditOutboxListProcessed возвращает SQL-запрос для получения списка успешно обработанных записей outbox
	// (внимание: SQL-запрос на данный момент не завершен).
	SqlDataAuditOutboxListProcessed string = `
select
    id,
    instance_id,
    content_type,
    payload,
    status,
    created_at,
    updated_at
from
    data_audit_outbox
where
    
`
	// SqlDataAuditOutboxChangeStatus возвращает SQL-запрос для частичного обновления только статуса записи outbox
	// и фиксации времени изменения.
	SqlDataAuditOutboxChangeStatus string = `
update
    data_audit_outbox
set
    status = $2,
    updated_at = $3
where
    id = $1
`
)
