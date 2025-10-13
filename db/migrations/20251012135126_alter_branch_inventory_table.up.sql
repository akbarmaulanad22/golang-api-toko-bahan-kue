ALTER TABLE branch_inventory
DROP FOREIGN KEY fk_branch_inventory_branch;

ALTER TABLE branch_inventory
DROP FOREIGN KEY fk_branch_inventory_size;

ALTER TABLE branch_inventory
ADD CONSTRAINT fk_branch_inventory_branch
FOREIGN KEY (branch_id)
REFERENCES branches(id)
ON DELETE CASCADE
ON UPDATE CASCADE,
ADD CONSTRAINT fk_branch_inventory_size
FOREIGN KEY (size_id)
REFERENCES sizes(id)
ON DELETE CASCADE
ON UPDATE CASCADE;
