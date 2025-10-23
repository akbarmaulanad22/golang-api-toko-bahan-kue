-- 1️⃣ Hapus FK lama yang masih nempel
ALTER TABLE purchase_details DROP FOREIGN KEY purchase_details_ibfk_1;
ALTER TABLE purchase_details DROP FOREIGN KEY purchase_details_ibfk_2;

-- 2️⃣ Hapus primary key komposit lama
ALTER TABLE purchase_details DROP PRIMARY KEY;

-- 3️⃣ Tambahkan kolom id (tanpa AUTO_INCREMENT dulu)
ALTER TABLE purchase_details ADD COLUMN id BIGINT NOT NULL FIRST;

-- 4️⃣ Jadikan kolom id sebagai PRIMARY KEY dan AUTO_INCREMENT
ALTER TABLE purchase_details MODIFY COLUMN id BIGINT NOT NULL AUTO_INCREMENT, ADD PRIMARY KEY (id);

-- 5️⃣ Tambahkan kembali unique constraint lama
ALTER TABLE purchase_details
ADD CONSTRAINT uq_purchase_details_code_size UNIQUE (purchase_code, size_id);
