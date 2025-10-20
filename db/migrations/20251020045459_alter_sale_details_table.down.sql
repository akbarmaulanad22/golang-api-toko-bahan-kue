ALTER TABLE sale_details
DROP PRIMARY KEY,
DROP COLUMN id,
DROP INDEX uq_sale_details_code_size,
ADD PRIMARY KEY (sale_code, size_id);
