-- 1️⃣ Hapus FK lama yang masih nempel
ALTER TABLE sale_details DROP FOREIGN KEY sale_details_ibfk_1;
ALTER TABLE sale_details DROP FOREIGN KEY sale_details_ibfk_2;

-- 2️⃣ Hapus primary key komposit lama
ALTER TABLE sale_details DROP PRIMARY KEY;

-- 3️⃣ Tambahkan kolom id (tanpa AUTO_INCREMENT dulu)
ALTER TABLE sale_details ADD COLUMN id BIGINT NOT NULL FIRST;

-- 4️⃣ Jadikan kolom id sebagai PRIMARY KEY dan AUTO_INCREMENT
ALTER TABLE sale_details MODIFY COLUMN id BIGINT NOT NULL AUTO_INCREMENT, ADD PRIMARY KEY (id);

-- 5️⃣ Tambahkan kembali unique constraint lama
ALTER TABLE sale_details
ADD CONSTRAINT uq_sale_details_code_size UNIQUE (sale_code, size_id);
