package example

const (
	sqlCreateTableTest2 string = `
create table if not exists test2 (
    id varchar(50) not null,
    code varchar(50) not null,
    name varchar(100) null,
    description varchar(512) null,
    created_at  timestamptz not null default now(),
    modified_at timestamptz not null default now(),
    constraint test2_pk primary key (id),
    constraint test2_uk unique (code)
)
`
	sqlDropTableTest2 string = `
drop table if exists test2 cascade;
`
	sqlCreateIndexTest2Code string = `create index if not exists idx_test2_code on test2 (code)`
	sqlDropIndexTest2Code   string = `drop index if exists idx_test2_code on test2 (code)`
)
