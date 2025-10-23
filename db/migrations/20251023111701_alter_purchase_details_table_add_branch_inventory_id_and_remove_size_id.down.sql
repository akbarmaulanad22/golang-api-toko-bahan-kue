-- 1️⃣ Hapus foreign key & unique constraint yang terkait dengan branch_inventory_id
ALTER TABLE purchase_details
DROP FOREIGN KEY fk_purchase_details_branch_inventory;

ALTER TABLE purchase_details
DROP INDEX uq_purchase_details_code_branchinv;

-- 2️⃣ Tambahkan kembali kolom size_id
ALTER TABLE purchase_details
ADD COLUMN size_id INT AFTER purchase_code;

-- 3️⃣ Isi kembali nilai size_id berdasarkan branch_inventory
UPDATE purchase_details sd
JOIN branch_inventory bi ON bi.id = sd.branch_inventory_id
SET sd.size_id = bi.size_id
WHERE sd.size_id IS NULL;

-- 4️⃣ Tambahkan kembali unique constraint lama
ALTER TABLE purchase_details
ADD CONSTRAINT uq_purchase_details_code_size UNIQUE (purchase_code, size_id);

-- 5️⃣ Hapus kolom branch_inventory_id
ALTER TABLE purchase_details
DROP COLUMN branch_inventory_id;
