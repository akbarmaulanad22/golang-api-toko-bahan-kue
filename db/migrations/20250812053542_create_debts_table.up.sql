CREATE TABLE debts (
    id                  INT AUTO_INCREMENT      PRIMARY KEY,
    reference_type      ENUM('SALE','PURCHASE') NOT NULL, 
    reference_code      VARCHAR(100)            NOT NULL,             
    total_amount        DECIMAL(15,2)           NOT NULL,              
    paid_amount         DECIMAL(15,2)           NOT NULL DEFAULT 0,      
    due_date            BIGINT NOT NULL,                          
    status              ENUM('PENDING','PAID', 'VOID')  DEFAULT 'PENDING',
    created_at          BIGINT                  NOT NULL,
    updated_at          BIGINT                  NOT NULL,
    FOREIGN KEY (reference_code) REFERENCES sales(code) ON UPDATE CASCADE ON DELETE CASCADE 
);
