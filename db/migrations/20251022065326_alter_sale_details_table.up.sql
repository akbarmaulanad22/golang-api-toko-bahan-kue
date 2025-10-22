-- 1️⃣ Drop unique constraint lama (sale_code, size_id)
ALTER TABLE sale_details
DROP INDEX uq_sale_details_code_size;

-- 2️⃣ Tambahkan kolom branch_inventory_id
ALTER TABLE sale_details
ADD COLUMN branch_inventory_id BIGINT AFTER sale_code;

-- 3️⃣ Tambahkan foreign key baru ke branch_inventory
ALTER TABLE sale_details
ADD CONSTRAINT fk_sale_details_branch_inventory
FOREIGN KEY (branch_inventory_id)
REFERENCES branch_inventory(id)
ON UPDATE CASCADE
ON DELETE CASCADE;

-- 4️⃣ Hapus kolom size_id
ALTER TABLE sale_details
DROP COLUMN size_id;

-- 5️⃣ Tambahkan unique constraint baru (sale_code, branch_inventory_id)
ALTER TABLE sale_details
ADD CONSTRAINT uq_sale_details_code_branchinv UNIQUE (sale_code, branch_inventory_id);
