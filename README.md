# eportalgo Backend

A Go backend project using `sqlc` for type-safe database access.

## Setup

1.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

2.  **Generate Go code from SQL:**
    ```bash
    sqlc generate
    ```

3.  **Run the application:**
    ```bash
    go run main.go
    ```

## Project Structure

- `db/schema/`: Database schema definitions (SQL).
- `db/queries/`: SQL queries.
- `db/`: Generated Go code for database access.
- `main.go`: Application entry point.
- `sqlc.yaml`: Configuration for `sqlc`.
- `.env`: Environment variables (do not commit!).
