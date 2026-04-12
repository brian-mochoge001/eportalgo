# Build stage
FROM golang:1.26-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# Use -ldflags to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# Run stage
FROM alpine:3.18

# Install ca-certificates (needed for Firebase and external APIs)
# Install postgresql-client for schema and seed scripts if needed
RUN apk add --no-cache ca-certificates postgresql-client bash

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy db folder for schema and seed scripts
COPY db/schema ./db/schema
COPY db/seed.sql ./db/seed.sql

# Copy entrypoint script
COPY scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Render injects the PORT environment variable.
# We expose 8080 as a default, but our app should listen on $PORT.
EXPOSE 8080

# The entrypoint script will run migrations/seed if needed and then start the app
ENTRYPOINT ["./entrypoint.sh"]
