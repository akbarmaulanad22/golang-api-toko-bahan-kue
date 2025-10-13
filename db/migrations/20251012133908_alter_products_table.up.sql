ALTER TABLE products
ADD CONSTRAINT fk_products_category
FOREIGN KEY (category_id)
REFERENCES categories(id)
ON DELETE CASCADE
ON UPDATE CASCADE;
