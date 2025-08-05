create table sale_details
(
    qty             int(11)                                   not null,
    is_cancelled    tinyint(1)                                not null,
    cancelled_at    bigint                                    null,
    
    sale_code       varchar(100)                              not null,
    size_id         int(11)                                   not null,
    
    primary key ( sale_code, size_id ),
    constraint `constr_sale_details_sale_code_fk`
        FOREIGN KEY `sale_fk` (`sale_code`) REFERENCES `sales` (`code`)
        ON DELETE CASCADE ON UPDATE CASCADE,
    constraint `constr_sale_details_size_id_fk`
        FOREIGN KEY `size_fk` (`size_id`) REFERENCES `sizes` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) engine = InnoDB;