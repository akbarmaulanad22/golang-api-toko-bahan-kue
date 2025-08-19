CREATE TABLE capitals (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    branch_id INT NOT NULL,
    type ENUM('IN','OUT') NOT NULL, -- IN = setoran modal, OUT = penarikan modal
    amount DECIMAL(15,2) NOT NULL,
    note VARCHAR(255),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);