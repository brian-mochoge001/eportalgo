#!/bin/bash
set -e

# Function to wait for postgres to be ready
wait_for_postgres() {
  echo "Waiting for PostgreSQL at $DATABASE_URL..."
  until psql "$DATABASE_URL" -c '\q' > /dev/null 2>&1; do
    echo "PostgreSQL is unavailable - sleeping..."
    sleep 2
  done
  echo "PostgreSQL is up - executing commands"
}

# If DATABASE_URL is set, wait for it
if [ -n "$DATABASE_URL" ]; then
  wait_for_postgres
fi

# Apply Schema files in order
# Railway and VPS often persist data, so we check if tables exist or use IF NOT EXISTS
# Since our schema files use CREATE TABLE (without IF NOT EXISTS), we might want to check
# if the database is already initialized.
# For simplicity in this seed script, we'll run them but you might see errors if tables exist.
# Recommended: Update your schema files to use CREATE TABLE IF NOT EXISTS.

echo "Applying schema..."
for f in db/schema/*.sql; do
  echo "Running $f..."
  psql "$DATABASE_URL" -f "$f" > /dev/null
done

echo "Running seed script..."
psql "$DATABASE_URL" -f db/seed.sql > /dev/null

echo "Starting backend..."
./main
