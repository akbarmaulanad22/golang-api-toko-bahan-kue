create table purchases
(
    code            varchar(100)                              not null,
    sales_name   varchar(100)                              not null,
    status          ENUM('PENDING', 'COMPLETED', 'CANCELLED') not null,
    cash_value      int(11)                                   not null,
    debit_value     int(11)                                   not null,
    paid_at         bigint                                    null,
    created_at      bigint                                    not null,
    cancelled_at    bigint                                    null,

    branch_id       int(11)                                   not null,
    distributor_id  int(11)                                   not null,
    
    primary key ( code ),
    foreign key ( branch_id )      references branches ( id ),
    foreign key ( distributor_id ) references distributors ( id )
) engine = InnoDB;