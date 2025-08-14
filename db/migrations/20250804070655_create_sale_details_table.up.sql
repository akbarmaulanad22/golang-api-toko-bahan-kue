CREATE TABLE sale_details (
    sale_code       VARCHAR(100)    NOT NULL,
    size_id         INT             NOT NULL,
    qty             INT             NOT NULL,
    sell_price      DECIMAL(15,2)   NOT NULL,
    is_cancelled    BOOLEAN         NOT NULL    DEFAULT FALSE,
    PRIMARY KEY (sale_code, size_id),
    FOREIGN KEY (sale_code) REFERENCES sales(code) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (size_id) REFERENCES sizes(id) ON UPDATE CASCADE ON DELETE CASCADE
);