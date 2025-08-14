CREATE TABLE sales (
    code VARCHAR(100) PRIMARY KEY,
    customer_name VARCHAR(100) NOT NULL,
    status ENUM('COMPLETED','CANCELLED') DEFAULT 'COMPLETED',
    created_at BIGINT NOT NULL,
    branch_id INT NOT NULL,
    FOREIGN KEY (branch_id) REFERENCES branches(id)
);