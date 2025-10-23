-- 1️⃣ Hapus AUTO_INCREMENT dari kolom id dulu
ALTER TABLE purchase_details
MODIFY COLUMN id BIGINT;

-- 2️⃣ Hapus primary key id
ALTER TABLE purchase_details
DROP PRIMARY KEY;

-- 3️⃣ Hapus kolom id
ALTER TABLE purchase_details
DROP COLUMN id;

-- 4️⃣ Tambahkan kembali primary key lama
ALTER TABLE purchase_details
ADD PRIMARY KEY (purchase_code, size_id);

-- 5️⃣ Tambahkan kembali foreign key lama
ALTER TABLE purchase_details
ADD CONSTRAINT purchase_details_ibfk_1
FOREIGN KEY (purchase_code) REFERENCES purchases(code)
ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE purchase_details
ADD CONSTRAINT purchase_details_ibfk_2
FOREIGN KEY (size_id) REFERENCES sizes(id)
ON UPDATE CASCADE ON DELETE CASCADE;
