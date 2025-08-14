CREATE TABLE purchase_details (
    purchase_code               VARCHAR(100)    NOT NULL,
    size_id                     INT             NOT NULL,
    qty                         INT             NOT NULL,
    buy_price                   DECIMAL(15,2)   NOT NULL,
    is_cancelled                BOOLEAN         NOT NULL    DEFAULT FALSE,
    PRIMARY KEY (purchase_code, size_id),
    FOREIGN KEY (purchase_code) REFERENCES purchases(code) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (size_id) REFERENCES sizes(id) ON UPDATE CASCADE ON DELETE CASCADE
);