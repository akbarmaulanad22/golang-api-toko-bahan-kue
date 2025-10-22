-- 1️⃣ Hapus foreign key ke branch_inventory
ALTER TABLE sale_details
DROP FOREIGN KEY fk_sale_details_branch_inventory;

-- 2️⃣ Hapus unique constraint baru (sale_code, branch_inventory_id)
ALTER TABLE sale_details
DROP INDEX uq_sale_details_code_branchinv;

-- 3️⃣ Tambahkan kembali kolom size_id
ALTER TABLE sale_details
ADD COLUMN size_id INT AFTER sale_code;

-- 4️⃣ Tambahkan kembali foreign key ke sizes
ALTER TABLE sale_details
ADD CONSTRAINT fk_sale_details_size
FOREIGN KEY (size_id)
REFERENCES sizes(id)
ON UPDATE CASCADE
ON DELETE CASCADE;

-- 5️⃣ Hapus kolom branch_inventory_id
ALTER TABLE sale_details
DROP COLUMN branch_inventory_id;

-- 6️⃣ Tambahkan kembali unique constraint lama (sale_code, size_id)
ALTER TABLE sale_details
ADD CONSTRAINT uq_sale_details_code_size UNIQUE (sale_code, size_id);
