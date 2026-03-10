#!/usr/bin/env bash
# Test script for build.sh
# Validates that the build script meets all requirements

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

test_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((TESTS_PASSED++))
}

test_fail() {
    echo -e "${RED}✗${NC} $1"
    ((TESTS_FAILED++))
}

echo "Testing build.sh..."
echo ""

# Test 1: Script is executable
if [[ -x build.sh ]]; then
    test_pass "Script is executable"
else
    test_fail "Script is not executable"
fi

# Test 2: Help flag works
if ./build.sh --help > /dev/null 2>&1; then
    test_pass "Help flag works"
else
    test_fail "Help flag failed"
fi

# Test 3: Development build
echo "Running development build..."
if ./build.sh --clean --mode dev > /dev/null 2>&1; then
    test_pass "Development build succeeded"
    if [[ -f bin/blacksector ]] || [[ -f bin/blacksector.exe ]]; then
        test_pass "Development binary created"
    else
        test_fail "Development binary not found"
    fi
else
    test_fail "Development build failed"
fi

# Test 4: Production build
echo "Running production build..."
if ./build.sh --mode prod > /dev/null 2>&1; then
    test_pass "Production build succeeded"
    
    # Check binary size (production should be smaller or similar)
    if [[ -f bin/blacksector ]]; then
        PROD_SIZE=$(stat -f%z bin/blacksector 2>/dev/null || stat -c%s bin/blacksector 2>/dev/null)
        if [[ $PROD_SIZE -gt 0 ]]; then
            test_pass "Production binary has valid size"
        else
            test_fail "Production binary size is invalid"
        fi
    fi
else
    test_fail "Production build failed"
fi

# Test 5: Version flag in binary
if [[ -f bin/blacksector ]]; then
    if ./bin/blacksector --version > /dev/null 2>&1; then
        test_pass "Binary version flag works"
        
        # Check version output contains expected fields
        VERSION_OUTPUT=$(./bin/blacksector --version)
        if echo "$VERSION_OUTPUT" | grep -q "Version:" && \
           echo "$VERSION_OUTPUT" | grep -q "Build Time:" && \
           echo "$VERSION_OUTPUT" | grep -q "Git Commit:"; then
            test_pass "Version output contains all required fields"
        else
            test_fail "Version output missing required fields"
        fi
    else
        test_fail "Binary version flag failed"
    fi
fi

# Test 6: Cross-compilation (Linux amd64)
echo "Testing cross-compilation..."
if ./build.sh --mode prod --target linux/amd64 > /dev/null 2>&1; then
    test_pass "Cross-compilation to linux/amd64 succeeded"
    if [[ -f bin/blacksector-linux-amd64 ]]; then
        test_pass "Cross-compiled binary created"
    else
        test_fail "Cross-compiled binary not found"
    fi
else
    test_fail "Cross-compilation failed"
fi

# Test 7: Clean flag
if ./build.sh --clean --mode dev > /dev/null 2>&1; then
    test_pass "Clean flag works"
else
    test_fail "Clean flag failed"
fi

# Test 8: Custom version string
if ./build.sh --mode dev --version "test-1.0.0" > /dev/null 2>&1; then
    test_pass "Custom version string works"
    if [[ -f bin/blacksector ]]; then
        VERSION_OUTPUT=$(./bin/blacksector --version)
        if echo "$VERSION_OUTPUT" | grep -q "test-1.0.0"; then
            test_pass "Custom version string injected correctly"
        else
            test_fail "Custom version string not found in binary"
        fi
    fi
else
    test_fail "Custom version string failed"
fi

# Summary
echo ""
echo "================================"
echo "Test Results:"
echo "  Passed: $TESTS_PASSED"
echo "  Failed: $TESTS_FAILED"
echo "================================"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
