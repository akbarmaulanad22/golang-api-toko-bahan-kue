create table distributors
(
    id         int(11) not null auto_increment,
    name       varchar(100) not null,
    address    varchar(100) not null,
    created_at bigint       not null,
    updated_at bigint       not null,
    primary key ( id ),
    unique ( name ),
    unique ( address )
) engine = InnoDB;