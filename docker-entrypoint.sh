#!/bin/sh

# Docker entrypoint script for Frontol ETL

set -e

# Function to wait for database
wait_for_db() {
    echo "Waiting for database to be ready..."
    until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
        echo "Database is unavailable - sleeping"
        sleep 2
    done
    echo "Database is ready!"
}

# Function to wait for FTP server
wait_for_ftp() {
    echo "Waiting for FTP server to be ready..."
    until nc -z "$FTP_HOST" "$FTP_PORT"; do
        echo "FTP server is unavailable - sleeping"
        sleep 2
    done
    echo "FTP server is ready!"
}

# Function to initialize database
# Note: Database initialization is done through migrations, not kassa_ddl.sql
init_database() {
    echo "Database initialization is handled by migrations."
    echo "Run './migrate up' to apply migrations."
}

# Function to create FTP directories
setup_ftp_dirs() {
    echo "Setting up FTP directories..."
    # This would typically be done by the FTP server container
    # but we can create local directories for testing
    mkdir -p "$LOCAL_DIR"/request "$LOCAL_DIR"/response
    # Ensure directory exists and is writable
    if [ -d "$LOCAL_DIR" ]; then
        chmod -R 755 "$LOCAL_DIR" 2>/dev/null || true
    fi
}

# Main execution
case "$1" in
    "webhook-server")
        wait_for_db
        wait_for_ftp
        setup_ftp_dirs
        echo "Starting webhook server..."
        exec ./webhook-server
        ;;
    "loader")
        wait_for_db
        wait_for_ftp
        setup_ftp_dirs
        echo "Starting loader..."
        exec ./frontol-loader
        ;;
    "loader-local")
        echo "Starting local file loader..."
        exec ./frontol-loader-local "$2"
        ;;
    "parser-test")
        echo "Starting parser test..."
        exec ./parser-test "$2"
        ;;
    "send-request")
        wait_for_ftp
        echo "Sending requests..."
        exec ./send-request
        ;;
    "clear-requests")
        wait_for_ftp
        echo "Clearing requests..."
        exec ./clear-requests
        ;;
    "init-db")
        wait_for_db
        init_database
        ;;
    "health-check")
        # Simple health check
        if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
            echo "Database: OK"
        else
            echo "Database: FAIL"
            exit 1
        fi
        
        if nc -z "$FTP_HOST" "$FTP_PORT" >/dev/null 2>&1; then
            echo "FTP: OK"
        else
            echo "FTP: FAIL"
            exit 1
        fi
        
        echo "All services: OK"
        ;;
    *)
        echo "Usage: $0 {webhook-server|loader|loader-local|parser-test|send-request|clear-requests|init-db|health-check}"
        echo ""
        echo "Commands:"
        echo "  webhook-server  - Start the webhook server"
        echo "  loader          - Run the ETL loader"
        echo "  loader-local    - Process local file (requires file path as second arg)"
        echo "  parser-test     - Test file parsing (requires file path as second arg)"
        echo "  send-request    - Send request files to FTP"
        echo "  clear-requests  - Clear FTP request/response folders"
        echo "  init-db         - Initialize database schema"
        echo "  health-check    - Check service health"
        exit 1
        ;;
esac
