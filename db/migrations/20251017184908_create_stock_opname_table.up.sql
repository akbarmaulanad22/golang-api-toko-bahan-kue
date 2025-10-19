CREATE TABLE stock_opname (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  branch_id BIGINT NOT NULL,
  date BIGINT NOT NULL,
  status ENUM('draft','completed','cancelled') DEFAULT 'draft',
  created_by varchar(255),
  verified_by varchar(255),
  created_at BIGINT,
  completed_at BIGINT
);
