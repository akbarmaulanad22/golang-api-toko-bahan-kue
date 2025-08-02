create table products
(
    sku              varchar(100) not null,
    name             varchar(100) not null,
    category_slug    varchar(100) not null,
    created_at       bigint       not null,
    updated_at       bigint       not null,
    primary key ( sku ),
    unique ( name ),
    foreign key ( category_slug ) references categories ( slug )
) engine = InnoDB;