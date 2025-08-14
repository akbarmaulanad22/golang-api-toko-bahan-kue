CREATE TABLE sale_payments (
    id              INT AUTO_INCREMENT  PRIMARY KEY,
    sale_code       VARCHAR(100)        NOT NULL,
    payment_method  ENUM('CASH', 'DEBIT', 'TRANSFER', 'QRIS', 'EWALLET') NOT NULL,
    amount          DECIMAL(15,2)       NOT NULL,
    note            VARCHAR(255),
    created_at      BIGINT              NOT NULL,
    FOREIGN KEY (sale_code) REFERENCES sales(code) ON UPDATE CASCADE ON DELETE CASCADE
);
