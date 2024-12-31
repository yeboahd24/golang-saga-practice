#!/bin/bash
set -e

# Wait for order-db
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "order-db" -U "postgres" -d "order_service" -c '\q'; do
  echo "order-db is unavailable - sleeping"
  sleep 1
done

echo "Initializing order-db..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "order-db" -U "postgres" -d "order_service" -f /init-scripts/init-order-db.sql

# Wait for payment-db
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "payment-db" -U "postgres" -d "payment_service" -c '\q'; do
  echo "payment-db is unavailable - sleeping"
  sleep 1
done

echo "Initializing payment-db..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "payment-db" -U "postgres" -d "payment_service" -f /init-scripts/init-payment-db.sql

# Wait for inventory-db
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "inventory-db" -U "postgres" -d "inventory_service" -c '\q'; do
  echo "inventory-db is unavailable - sleeping"
  sleep 1
done

echo "Initializing inventory-db..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "inventory-db" -U "postgres" -d "inventory_service" -f /init-scripts/init-inventory-db.sql

echo "All databases initialized successfully"
