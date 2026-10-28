package postgres

const (
	SqlTest2Find string = `
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
	SqlTest2FindByCode string = `
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
	SqlTest2List string = `
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
	SqlTest2Create string = `
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
	SqlTest2Change string = `
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
	SqlTest2Delete string = `
delete
from
    test2
where
    id = $1
`
)
