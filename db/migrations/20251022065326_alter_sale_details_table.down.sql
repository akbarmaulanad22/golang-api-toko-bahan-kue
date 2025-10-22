-- 1️⃣ Hapus foreign key & unique constraint yang terkait dengan branch_inventory_id
ALTER TABLE sale_details
DROP FOREIGN KEY fk_sale_details_branch_inventory;

ALTER TABLE sale_details
DROP INDEX uq_sale_details_code_branchinv;

-- 2️⃣ Tambahkan kembali kolom size_id
ALTER TABLE sale_details
ADD COLUMN size_id INT AFTER sale_code;

-- 3️⃣ Isi kembali nilai size_id berdasarkan branch_inventory
UPDATE sale_details sd
JOIN branch_inventory bi ON bi.id = sd.branch_inventory_id
SET sd.size_id = bi.size_id
WHERE sd.size_id IS NULL;

-- 4️⃣ Tambahkan kembali unique constraint lama
ALTER TABLE sale_details
ADD CONSTRAINT uq_sale_details_code_size UNIQUE (sale_code, size_id);

-- 5️⃣ Hapus kolom branch_inventory_id
ALTER TABLE sale_details
DROP COLUMN branch_inventory_id;
