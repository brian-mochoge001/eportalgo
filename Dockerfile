# Build stage
FROM golang:1.26.1-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:3.18

# Install postgresql-client for running seed scripts and schema
RUN apk add --no-cache postgresql-client bash

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy db folder for schema and seed scripts
COPY db/schema ./db/schema
COPY db/seed.sql ./db/seed.sql

# Copy entrypoint script
COPY scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Expose port
EXPOSE 8080

# Use entrypoint script to run migrations/seed and then start the app
ENTRYPOINT ["./entrypoint.sh"]
