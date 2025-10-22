-- 1️⃣ Hapus AUTO_INCREMENT dari kolom id dulu
ALTER TABLE sale_details
MODIFY COLUMN id BIGINT;

-- 2️⃣ Hapus primary key id
ALTER TABLE sale_details
DROP PRIMARY KEY;

-- 3️⃣ Hapus kolom id
ALTER TABLE sale_details
DROP COLUMN id;

-- 4️⃣ Tambahkan kembali primary key lama
ALTER TABLE sale_details
ADD PRIMARY KEY (sale_code, size_id);

-- 5️⃣ Tambahkan kembali foreign key lama
ALTER TABLE sale_details
ADD CONSTRAINT sale_details_ibfk_1
FOREIGN KEY (sale_code) REFERENCES sales(code)
ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE sale_details
ADD CONSTRAINT sale_details_ibfk_2
FOREIGN KEY (size_id) REFERENCES sizes(id)
ON UPDATE CASCADE ON DELETE CASCADE;
