ALTER TABLE purchase_details
DROP PRIMARY KEY,
ADD COLUMN id BIGINT AUTO_INCREMENT PRIMARY KEY FIRST,
ADD CONSTRAINT uq_purchase_details_code_size UNIQUE (purchase_code, size_id);
