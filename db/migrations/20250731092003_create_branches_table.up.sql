create table branches
(
    id         int(11) not null auto_increment,
    name       varchar(100) not null,
    address    varchar(100) not null,
    created_at bigint       not null,
    updated_at bigint       not null,
    primary key ( id ),
    constraint UC_BRANCHES UNIQUE ( name, address )
) engine = InnoDB;