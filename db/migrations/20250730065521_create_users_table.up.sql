create table users
(
    username   varchar(100) not null,
    password   varchar(100) not null,
    name       varchar(100) not null,
    address    varchar(100) not null,
    token      varchar(100) null,
    created_at bigint       not null,
    updated_at bigint       not null,
    primary key ( username ),
    unique ( username )
) engine = InnoDB;