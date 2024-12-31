#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS inventory (
        product_id VARCHAR(36) PRIMARY KEY,
        quantity INT NOT NULL,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    INSERT INTO inventory (product_id, quantity) VALUES
        ('prod1', 100),
        ('prod2', 100)
    ON CONFLICT (product_id) DO UPDATE SET quantity = EXCLUDED.quantity;
EOSQL
