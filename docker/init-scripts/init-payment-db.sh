#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS payments (
        id VARCHAR(36) PRIMARY KEY,
        order_id VARCHAR(36) NOT NULL,
        amount DECIMAL(10,2) NOT NULL,
        status VARCHAR(20) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
EOSQL
