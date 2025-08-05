create table categories
(
    id         int(11)      not null auto_increment,
    name       varchar(100) not null,
    created_at bigint       not null,
    updated_at bigint       not null,
    primary key ( id ),
    unique ( name )
) engine = InnoDB;