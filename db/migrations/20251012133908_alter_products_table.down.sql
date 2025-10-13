ALTER TABLE products
DROP FOREIGN KEY fk_products_category;

ALTER TABLE products
ADD CONSTRAINT products_ibfk_1
FOREIGN KEY (category_id)
REFERENCES categories(id);
