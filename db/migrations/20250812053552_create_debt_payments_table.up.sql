CREATE TABLE debt_payments (
    id              INT AUTO_INCREMENT  PRIMARY KEY,
    debt_id         INT                 NOT NULL,
    payment_date    BIGINT              NOT NULL,
    amount          DECIMAL(15,2)       NOT NULL,
    note            VARCHAR(255),
    created_at      BIGINT              NOT NULL,
    updated_at      BIGINT              NOT NULL,
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON UPDATE CASCADE ON DELETE CASCADE
);
