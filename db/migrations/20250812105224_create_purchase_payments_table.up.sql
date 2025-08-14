CREATE TABLE purchase_payments (
    id              INT                                                     AUTO_INCREMENT  PRIMARY KEY,
    purchase_code   VARCHAR(100)                                            NOT NULL,
    payment_method  ENUM('CASH', 'DEBIT', 'TRANSFER', 'QRIS', 'EWALLET')    NOT NULL,
    amount          DECIMAL(15,2)                                           NOT NULL,
    created_at      BIGINT                                                  NOT NULL,
    note            VARCHAR(255),
    FOREIGN KEY (purchase_code) REFERENCES purchases(code) ON UPDATE CASCADE ON DELETE CASCADE
);