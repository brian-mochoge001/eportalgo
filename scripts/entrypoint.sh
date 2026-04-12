#!/bin/bash
set -e

# Render and other managed environments provide DATABASE_URL.
# If INITIALIZE_DB=true is set in Render environment, we run schema and seed.
# Otherwise, we skip it to prevent accidental data loss or long startup times.

if [ "$INITIALIZE_DB" = "true" ]; then
    echo "Starting database initialization..."

    # Function to wait for postgres to be ready
    wait_for_postgres() {
        echo "Waiting for PostgreSQL at $DATABASE_URL..."
        until psql "$DATABASE_URL" -c '\q' > /dev/null 2>&1; do
            echo "PostgreSQL is unavailable - sleeping..."
            sleep 2
        done
        echo "PostgreSQL is up - executing commands"
    }

    if [ -n "$DATABASE_URL" ]; then
        wait_for_postgres
    fi

    echo "Applying schema..."
    # Using IF NOT EXISTS or checking for table presence is recommended in the SQL files themselves.
    # Here we just iterate and apply.
    for f in db/schema/*.sql; do
        echo "Running $f..."
        psql "$DATABASE_URL" -f "$f" > /dev/null
    done

    echo "Running seed script..."
    psql "$DATABASE_URL" -f db/seed.sql > /dev/null

    echo "Database initialization completed."
fi

# In Render, the application MUST listen on the port defined by $PORT.
# Our main.go handles this via os.Getenv("PORT").

echo "Starting backend..."
./main
