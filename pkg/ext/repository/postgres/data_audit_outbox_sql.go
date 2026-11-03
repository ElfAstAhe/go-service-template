package postgres

const (
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
	SqlDataAuditOutboxDelete string = `
delete
from
    data_audit_outbox
where
    id = $1
`
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
