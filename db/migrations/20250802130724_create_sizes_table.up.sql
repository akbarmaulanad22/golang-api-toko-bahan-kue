create table sizes
(
    id         int(11)      not null auto_increment,
    name       varchar(100) not null,
    sell_price int(11)      not null,
    buy_price  int(11)      not null,
    created_at bigint       not null,
    updated_at bigint       not null,

    product_sku varchar(100) not null,
    
    primary key ( id ),
    constraint uc_sizes UNIQUE (name, product_sku),
    foreign key ( product_sku ) references products ( sku )
) engine = InnoDB;