\c inventory_service;

BEGIN;

CREATE TABLE IF NOT EXISTS inventory (
    product_id VARCHAR(36) PRIMARY KEY,
    quantity INT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert some sample inventory data
INSERT INTO inventory (product_id, quantity) VALUES
    ('prod1', 100),
    ('prod2', 100)
ON CONFLICT (product_id) DO UPDATE SET quantity = EXCLUDED.quantity;

COMMIT;
