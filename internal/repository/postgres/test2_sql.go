package postgres

const (
	sqlTest2Find = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test2
where
    id = $1
`
	sqlTest2FindByCode = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test2
where
    code = $1
`
	sqlTest2List string = `
select
    id,
    code,
    name,
    description,
    created_at,
    modified_at
from
    test2
order by
    id asc
offset $2
limit $1
`
	sqlTest2Create = `
insert into test2 (
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
	sqlTest2Change = `
update
    test2
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
	sqlTest2Delete = `
delete
from
    test2
where
    id = $1
`
)
