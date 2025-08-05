create table sales
(
    code            varchar(100)                              not null,
    customer_name   varchar(100)                              not null,
    status          ENUM('PENDING', 'COMPLETED', 'CANCELLED') not null,
    cash_value      int(11)                                   not null,
    debit_value     int(11)                                   not null,
    paid_at         bigint                                    null,
    created_at      bigint                                    not null,
    cancelled_at    bigint                                    null,

    branch_id       int(11)                                   not null,
    
    primary key ( code ),
    foreign key ( branch_id )      references branches ( id )
) engine = InnoDB;