ALTER TABLE purchase_details
DROP PRIMARY KEY,
DROP COLUMN id,
DROP INDEX uq_purchase_details_code_size,
ADD PRIMARY KEY (purchase_code, size_id);
