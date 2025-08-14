create table expenses (
    id              int(11)         not null auto_increment,
    description     varchar(255)    not null,
    amount          int(11)         not null,
    branch_id       int(11)         not null,
    created_at      bigint          not null,
    updated_at      bigint          not null,
    primary key (id),
    foreign key (branch_id) references branches(id) on delete cascade on update cascade
) engine=InnoDB;