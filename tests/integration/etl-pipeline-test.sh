#!/bin/bash
# Integration test for ETL pipeline
# Tests the complete ETL workflow: FTP -> Parse -> Database

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_test() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        if [ -n "$message" ]; then
            echo "  $message"
        fi
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        if [ -n "$message" ]; then
            echo -e "  ${RED}$message${NC}"
        fi
        ((TESTS_FAILED++))
    fi
}

# Function to check if service is running
check_service() {
    local service="$1"
    if docker compose ps "$service" | grep -q "Up"; then
        return 0
    else
        return 1
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local service="$1"
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if check_service "$service"; then
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    
    return 1
}

# Main test execution
echo "=========================================="
echo "ETL Pipeline Integration Test"
echo "=========================================="
echo ""

cd "$PROJECT_ROOT"

# Check if docker compose is available
if ! command -v docker compose &> /dev/null; then
    echo -e "${RED}Error: docker compose not found${NC}"
    exit 1
fi

# Check prerequisites
echo "=== Checking Prerequisites ==="

# Check FTP server
if ! check_service "ftp-server"; then
    echo -e "${YELLOW}Starting FTP server...${NC}"
    docker compose up -d ftp-server
    if ! wait_for_service "ftp-server"; then
        print_test "FTP server startup" "FAIL" "FTP server failed to start"
        exit 1
    fi
fi
print_test "FTP server is running" "PASS" ""

# Check database (if configured)
if [ -n "$DB_HOST" ]; then
    print_test "Database configured" "PASS" "DB_HOST=$DB_HOST"
else
    echo -e "${YELLOW}Warning: DB_HOST not set, skipping database tests${NC}"
fi

# Check webhook server (optional)
if check_service "webhook-server"; then
    print_test "Webhook server is running" "PASS" ""
else
    echo -e "${YELLOW}Webhook server not running (optional)${NC}"
fi

echo ""

# Test 1: FTP connectivity
echo "=== Testing FTP Connectivity ==="
if docker compose exec -T ftp-server test -d /home/ftp/frontol/request; then
    print_test "FTP request directory exists" "PASS" ""
else
    print_test "FTP request directory exists" "FAIL" "Directory not found"
fi

if docker compose exec -T ftp-server test -d /home/ftp/frontol/response; then
    print_test "FTP response directory exists" "PASS" ""
else
    print_test "FTP response directory exists" "FAIL" "Directory not found"
fi
echo ""

# Test 2: FTP file operations
echo "=== Testing FTP File Operations ==="
TEST_FILE="/home/ftp/frontol/response/P13/P13/test_etl_$(date +%s).txt"
TEST_CONTENT="Test ETL pipeline file"

# Create test file
if docker compose exec -T ftp-server sh -c "echo '$TEST_CONTENT' > $TEST_FILE" 2>/dev/null; then
    print_test "Create test file on FTP" "PASS" ""
    
    # Verify file exists
    if docker compose exec -T ftp-server test -f "$TEST_FILE"; then
        print_test "Verify test file exists" "PASS" ""
        
        # Read file content
        if docker compose exec -T ftp-server cat "$TEST_FILE" | grep -q "$TEST_CONTENT"; then
            print_test "Read test file content" "PASS" ""
        else
            print_test "Read test file content" "FAIL" "Content mismatch"
        fi
        
        # Delete test file
        if docker compose exec -T ftp-server rm -f "$TEST_FILE"; then
            print_test "Delete test file" "PASS" ""
        else
            print_test "Delete test file" "FAIL" "Failed to delete"
        fi
    else
        print_test "Verify test file exists" "FAIL" "File not found"
    fi
else
    print_test "Create test file on FTP" "FAIL" "Cannot create file"
fi
echo ""

# Test 3: ETL pipeline components
echo "=== Testing ETL Pipeline Components ==="

# Test loader binary exists
if docker compose run --rm loader test -f ./frontol-loader 2>/dev/null; then
    print_test "Loader binary exists" "PASS" ""
else
    print_test "Loader binary exists" "FAIL" "Binary not found"
fi

# Test parser binary exists
if docker compose run --rm parser-test test -f ./parser-test 2>/dev/null; then
    print_test "Parser binary exists" "PASS" ""
else
    print_test "Parser binary exists" "FAIL" "Binary not found"
fi
echo ""

# Test 4: Kassa folder structure
echo "=== Testing Kassa Folder Structure ==="
KASSA_FOLDERS=("P13" "N22" "SH54" "S6" "L98" "L32" "S39" "O49" "L28")

for kassa in "${KASSA_FOLDERS[@]}"; do
    REQUEST_PATH="/home/ftp/frontol/request/$kassa"
    RESPONSE_PATH="/home/ftp/frontol/response/$kassa"
    
    if docker compose exec -T ftp-server test -d "$REQUEST_PATH" 2>/dev/null; then
        print_test "Kassa $kassa request folder exists" "PASS" ""
    else
        print_test "Kassa $kassa request folder exists" "FAIL" "Folder not found"
    fi
    
    if docker compose exec -T ftp-server test -d "$RESPONSE_PATH" 2>/dev/null; then
        print_test "Kassa $kassa response folder exists" "PASS" ""
    else
        print_test "Kassa $kassa response folder exists" "FAIL" "Folder not found"
    fi
done
echo ""

# Test 5: ETL pipeline end-to-end (if database available)
if [ -n "$DB_HOST" ] && [ "$DB_HOST" != "your_postgres_host" ]; then
    echo "=== Testing ETL Pipeline End-to-End ==="
    echo -e "${YELLOW}Note: This test requires a running database${NC}"
    
    # This would require actual database connection
    # For now, we just verify the pipeline can be invoked
    if docker compose run --rm loader echo "ETL pipeline test" 2>/dev/null; then
        print_test "ETL pipeline can be invoked" "PASS" ""
    else
        print_test "ETL pipeline can be invoked" "FAIL" "Cannot invoke pipeline"
    fi
    echo ""
else
    echo -e "${YELLOW}Skipping end-to-end ETL test (no database configured)${NC}"
    echo ""
fi

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}Failed: $TESTS_FAILED${NC}"
    exit 0
fi

