CREATE TABLE branch_inventory (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    branch_id int(11) NOT NULL,
    size_id int(11) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    UNIQUE KEY uq_branch_size (branch_id, size_id),
    CONSTRAINT fk_branch_inventory_branch FOREIGN KEY (branch_id) REFERENCES branches(id),
    CONSTRAINT fk_branch_inventory_size FOREIGN KEY (size_id) REFERENCES sizes(id)
);