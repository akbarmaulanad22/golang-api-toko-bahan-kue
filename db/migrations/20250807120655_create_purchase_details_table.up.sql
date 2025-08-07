create table purchase_details
(
    qty             int(11)                                   not null,
    is_cancelled    tinyint(1)                                not null,
    cancelled_at    bigint                                    null,
    
    purchase_code   varchar(100)                              not null,
    size_id         int(11)                                   not null,
    
    primary key ( purchase_code, size_id ),
    constraint `constr_purchase_details_purchase_code_fk`
        FOREIGN KEY `purchase_fk` (`purchase_code`) REFERENCES `purchases` (`code`)
        ON DELETE CASCADE ON UPDATE CASCADE,
    constraint `constr_purchase_details_size_id_fk`
        FOREIGN KEY `size_fk` (`size_id`) REFERENCES `sizes` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) engine = InnoDB;