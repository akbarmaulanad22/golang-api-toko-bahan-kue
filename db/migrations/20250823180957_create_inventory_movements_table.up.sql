CREATE TABLE inventory_movements (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    branch_inventory_id BIGINT NOT NULL,
    change_qty INT NOT NULL,
    reference_type VARCHAR(50) NOT NULL,
    reference_key VARCHAR(255),
    created_at BIGINT NOT NULL,
    CONSTRAINT fk_inventory_movements_inventory FOREIGN KEY (branch_inventory_id) REFERENCES branch_inventory(id)
);