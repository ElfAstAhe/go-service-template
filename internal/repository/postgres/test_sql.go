package postgres

const (
	sqlTestFind = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test
where
    id = $1
`
	sqlTestFindByCode = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test
where
    code = $1
`
	sqlTestList string = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test
order by
    id asc
offset $2
limit $1
`
	sqlTestCreate = `
insert into test (
    id,
    code,
    name,
    description,
    created_at,
    modified_at
)
values (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
)
returning
    id,
    code,
    name,
    description,
    created_at,
    modified_at
`
	sqlTestChange = `
update
    test
set
    code = $2,
    name = $3,
    description = $4,
    modified_at = $5
where
    id = $1
returning
    id,
    code,
    name,
    description,
    created_at,
    modified_at
`
	sqlTestDelete = `
delete
from
    test
where
    id = $1
`
)
