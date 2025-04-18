#!/bin/bash

# Run the application with explicit environment variables
DB_HOST="127.0.0.1" \
    DB_PORT="3306" \
    DB_USER="root" \
    DB_PASSWORD="root" \
    DB_NAME="eedb" \
    go run main.go
