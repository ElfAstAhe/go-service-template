package example

const (
	sqlCreateTableTest = `
create table if not exists test (
    id varchar(50) not null,
    code varchar(50) not null,
    name varchar(100) null,
    description varchar(512) null,
    created_at  timestamptz not null default now(),
    modified_at timestamptz not null default now(),
    constraint test_pk primary key (id),
    constraint test_uk unique (code)
)
`
	sqlDropTableTest = `
drop table if exists test
`
	sqlCreateIndexTestCode = `create index if not exists idx_test_code on test (code asc)`
	sqlDropIndexTestCode   = `drop index if exists idx_test_code`
)
