CREATE TABLE stock_opname_detail (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  stock_opname_id BIGINT NOT NULL,
  branch_inventory_id BIGINT NOT NULL,
  system_qty BIGINT NOT NULL,
  physical_qty BIGINT NOT NULL,
  difference BIGINT NOT NULL,
  notes TEXT,
  FOREIGN KEY (stock_opname_id) REFERENCES stock_opname(id) ON DELETE CASCADE ON UPDATE CASCADE
);
