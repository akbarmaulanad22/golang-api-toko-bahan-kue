create table users
(
    username   varchar(100) not null,
    password   varchar(100) not null,
    name       varchar(100) not null,
    address    varchar(100) not null,
    token      varchar(100) null,
    created_at bigint       not null,
    updated_at bigint       not null,

    role_id    int(11)      not null,
    branch_id  int(11)      not null,
    
    primary key ( username ),
    unique ( username ),
    foreign key ( role_id ) references roles ( id ) ON DELETE CASCADE ON UPDATE CASCADE,
    foreign key ( branch_id ) references branches ( id ) ON DELETE CASCADE ON UPDATE CASCADE
) engine = InnoDB;