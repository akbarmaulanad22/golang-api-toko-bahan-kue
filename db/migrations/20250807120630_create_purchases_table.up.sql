
CREATE TABLE purchases (
    distributor_id  INT                                 NOT NULL,
    branch_id       INT                                 NOT NULL,
    code            VARCHAR(100)                        PRIMARY KEY,
    sales_name      VARCHAR(100)                        NOT NULL,
    status          ENUM('COMPLETED','CANCELLED')       NOT NULL,
    created_at      BIGINT NOT NULL,
    FOREIGN KEY  (branch_id) REFERENCES branches(id),
    FOREIGN KEY  (distributor_id) REFERENCES distributors(id)
);