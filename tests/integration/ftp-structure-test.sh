#!/bin/bash
# Integration tests for FTP structure init container
# Tests folder creation, permissions, and file operations

# Don't exit on error - we want to run all tests
set +e

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

# Function to check if directory exists
test_directory_exists() {
    local path="$1"
    local test_name="$2"
    
    if docker compose exec ftp-server test -d "$path" >/dev/null 2>&1; then
        print_test "$test_name" "PASS" "Directory exists: $path"
        return 0
    else
        print_test "$test_name" "FAIL" "Directory does not exist: $path"
        return 1
    fi
}

# Function to check directory permissions
test_directory_permissions() {
    local path="$1"
    local expected_owner="$2"
    local expected_group="$3"
    local expected_mode="$4"
    local test_name="$5"
    
    local perms=$(docker compose exec ftp-server stat -c "%a %U %G" "$path" 2>/dev/null)
    local mode=$(echo "$perms" | cut -d' ' -f1)
    local owner=$(echo "$perms" | cut -d' ' -f2)
    local group=$(echo "$perms" | cut -d' ' -f3)
    
    local errors=()
    
    if [ "$owner" != "$expected_owner" ]; then
        errors+=("Owner mismatch: expected $expected_owner, got $owner")
    fi
    
    if [ "$group" != "$expected_group" ]; then
        errors+=("Group mismatch: expected $expected_group, got $group")
    fi
    
    if [ "$mode" != "$expected_mode" ]; then
        errors+=("Mode mismatch: expected $expected_mode, got $mode")
    fi
    
    if [ ${#errors[@]} -eq 0 ]; then
        print_test "$test_name" "PASS" "Permissions: $mode $owner:$group"
        return 0
    else
        print_test "$test_name" "FAIL" "$(IFS='; '; echo "${errors[*]}")"
        return 1
    fi
}

# Function to test file creation
test_file_creation() {
    local dir="$1"
    local test_name="$2"
    local test_file="$dir/ftp_test_$(date +%s).txt"
    
    if docker compose exec ftp-server su -s /bin/sh ftpuser -c "touch $test_file && test -f $test_file" 2>/dev/null; then
        docker compose exec ftp-server rm -f "$test_file" 2>/dev/null
        print_test "$test_name" "PASS" "File created and removed successfully"
        return 0
    else
        print_test "$test_name" "FAIL" "Cannot create file in $dir"
        return 1
    fi
}

# Function to test recursive permissions
test_recursive_permissions() {
    local base_dir="$1"
    local expected_owner="$2"
    local expected_group="$3"
    local test_name="$4"
    
    local dirs=$(docker compose exec ftp-server find "$base_dir" -type d 2>/dev/null)
    local failed=0
    local total=0
    
    while IFS= read -r dir; do
        [ -z "$dir" ] && continue
        ((total++))
        local perms=$(docker compose exec ftp-server stat -c "%a %U %G" "$dir" 2>/dev/null)
        local owner=$(echo "$perms" | cut -d' ' -f2)
        local group=$(echo "$perms" | cut -d' ' -f3)
        
        if [ "$owner" != "$expected_owner" ] || [ "$group" != "$expected_group" ]; then
            ((failed++))
        fi
    done <<< "$dirs"
    
    if [ $failed -eq 0 ]; then
        print_test "$test_name" "PASS" "All $total directories have correct ownership"
        return 0
    else
        print_test "$test_name" "FAIL" "$failed of $total directories have incorrect ownership"
        return 1
    fi
}

# Function to test idempotency
test_idempotency() {
    local test_name="Idempotency test"
    
    # Get initial directory count
    local initial_count=$(docker compose exec ftp-server find /home/ftp/frontol/request -type d 2>/dev/null | wc -l)
    
    # Run init container again
    docker compose run --rm ftp-structure-init > /dev/null 2>&1
    
    # Get directory count after rerun
    local final_count=$(docker compose exec ftp-server find /home/ftp/frontol/request -type d 2>/dev/null | wc -l)
    
    if [ "$initial_count" -eq "$final_count" ]; then
        print_test "$test_name" "PASS" "Directory count unchanged ($initial_count)"
        return 0
    else
        print_test "$test_name" "FAIL" "Directory count changed: $initial_count -> $final_count"
        return 1
    fi
}

# Main test execution
echo "=========================================="
echo "FTP Structure Init Container Tests"
echo "=========================================="
echo ""

cd "$PROJECT_ROOT"

# Check if docker compose is available
if ! command -v docker compose &> /dev/null; then
    echo -e "${RED}Error: docker compose not found${NC}"
    exit 1
fi

# Check if FTP server is running
if ! docker compose ps ftp-server | grep -q "Up"; then
    echo -e "${YELLOW}Warning: FTP server is not running. Starting services...${NC}"
    docker compose up -d ftp-server > /dev/null 2>&1
    sleep 3
fi

# Test 1: Base directories exist
echo "=== Testing Base Directories ==="
test_directory_exists "/home/ftp/frontol/request" "Base request directory exists"
test_directory_exists "/home/ftp/frontol/response" "Base response directory exists"
echo ""

# Test 2: Base directory permissions
echo "=== Testing Base Directory Permissions ==="
test_directory_permissions "/home/ftp/frontol/request" "ftpuser" "ftpgroup" "775" "Request directory permissions"
test_directory_permissions "/home/ftp/frontol/response" "ftpuser" "ftpgroup" "775" "Response directory permissions"
echo ""

# Test 3: Kassa folders exist (based on default KASSA_STRUCTURE)
echo "=== Testing Kassa Folder Structure ==="
test_directory_exists "/home/ftp/frontol/request/P13" "Kassa P13 request folder exists"
test_directory_exists "/home/ftp/frontol/request/P13/P13" "Kassa P13/P13 request folder exists"
test_directory_exists "/home/ftp/frontol/response/P13/P13" "Kassa P13/P13 response folder exists"
echo ""

# Test 4: Kassa folder permissions
echo "=== Testing Kassa Folder Permissions ==="
test_directory_permissions "/home/ftp/frontol/request/P13/P13" "ftpuser" "ftpgroup" "775" "P13/P13 request folder permissions"
test_directory_permissions "/home/ftp/frontol/response/P13/P13" "ftpuser" "ftpgroup" "775" "P13/P13 response folder permissions"
echo ""

# Test 5: File creation in request folders
echo "=== Testing File Creation ==="
test_file_creation "/home/ftp/frontol/request/P13/P13" "File creation in request folder"
test_file_creation "/home/ftp/frontol/response/P13/P13" "File creation in response folder"
echo ""

# Test 6: Recursive permissions
echo "=== Testing Recursive Permissions ==="
test_recursive_permissions "/home/ftp/frontol/request" "ftpuser" "ftpgroup" "Request directory recursive permissions"
test_recursive_permissions "/home/ftp/frontol/response" "ftpuser" "ftpgroup" "Response directory recursive permissions"
echo ""

# Test 7: Idempotency
echo "=== Testing Idempotency ==="
test_idempotency
echo ""

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

