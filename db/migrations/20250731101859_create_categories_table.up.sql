create table categories
(
    slug       varchar(100) not null,
    name       varchar(100) not null,
    created_at bigint       not null,
    updated_at bigint       not null,
    primary key ( slug ),
    unique ( name )
) engine = InnoDB;