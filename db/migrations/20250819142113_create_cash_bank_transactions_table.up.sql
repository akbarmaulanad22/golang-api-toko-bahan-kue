CREATE TABLE cash_bank_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    branch_id INT NOT NULL,
    transaction_date BIGINT NOT NULL,
    type ENUM('IN','OUT') NOT NULL,
    source ENUM('SALE','PURCHASE','EXPENSE','DEBT','CAPITAL') NOT NULL,
    reference_key VARCHAR(100), -- kode transaksi sumber (misal sale_code, purchase_code, capital_id)
    amount DECIMAL(15,2) NOT NULL,
    description VARCHAR(255),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);
