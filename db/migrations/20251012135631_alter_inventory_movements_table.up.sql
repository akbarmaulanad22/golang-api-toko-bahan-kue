ALTER TABLE inventory_movements
DROP FOREIGN KEY fk_inventory_movements_inventory;

ALTER TABLE inventory_movements
ADD CONSTRAINT fk_inventory_movements_inventory
FOREIGN KEY (branch_inventory_id)
REFERENCES branch_inventory(id)
ON DELETE CASCADE
ON UPDATE CASCADE;
