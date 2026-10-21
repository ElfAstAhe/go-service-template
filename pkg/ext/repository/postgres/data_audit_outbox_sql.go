package postgres

const (
	sqlDataAuditOutboxFind string = `
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
	sqlDataAuditOutboxList string = `
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
	sqlDataAuditOutboxCreate string = `
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
	sqlDataAuditOutboxChange string = `
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
	sqlDataAuditOutboxDelete string = `
delete
from
    data_audit_outbox
where
    id = $1
`
	sqlDataAuditOutboxGetTail string = `
select
    id
from
    data_audit_aoutbox
where
    status in ('SUCCESS', 'FAILED')
and updated_at < $1
`
)
