#!/bin/bash

# staticSend Test Runner Script
# This script provides a simple way to run different types of tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    if go test ./pkg/... -v -cover; then
        print_success "Unit tests passed!"
        return 0
    else
        print_error "Unit tests failed!"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    if go test -v ./integration_test.go; then
        print_success "Integration tests passed!"
        return 0
    else
        print_error "Integration tests failed!"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests..."
    
    unit_result=0
    integration_result=0
    
    # Run unit tests
    if ! run_unit_tests; then
        unit_result=1
    fi
    
    echo ""
    
    # Run integration tests
    if ! run_integration_tests; then
        integration_result=1
    fi
    
    echo ""
    
    # Summary
    if [ $unit_result -eq 0 ] && [ $integration_result -eq 0 ]; then
        print_success "All tests passed! ✅"
        return 0
    else
        if [ $unit_result -ne 0 ]; then
            print_error "Unit tests failed ❌"
        fi
        if [ $integration_result -ne 0 ]; then
            print_error "Integration tests failed ❌"
        fi
        return 1
    fi
}

# Function to run tests with coverage
run_coverage() {
    print_status "Running tests with coverage report..."
    
    # Create coverage directory if it doesn't exist
    mkdir -p coverage
    
    # Run unit tests with coverage
    if go test ./pkg/... -coverprofile=coverage/coverage.out; then
        # Generate HTML coverage report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        print_success "Coverage report generated: coverage/coverage.html"
        
        # Show coverage summary
        print_status "Coverage summary:"
        go tool cover -func=coverage/coverage.out | tail -1
        
        return 0
    else
        print_error "Coverage tests failed!"
        return 1
    fi
}

# Function to show help
show_help() {
    echo "staticSend Test Runner"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit         Run unit tests only"
    echo "  integration  Run integration tests only"
    echo "  all          Run all tests (default)"
    echo "  coverage     Run tests with coverage report"
    echo "  help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all tests"
    echo "  $0 unit         # Run unit tests only"
    echo "  $0 integration  # Run integration tests only"
    echo "  $0 coverage     # Run with coverage report"
}

# Main script logic
main() {
    case "${1:-all}" in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "all")
            run_all_tests
            ;;
        "coverage")
            run_coverage
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Change to script directory
cd "$(dirname "$0")/.."

# Run main function with all arguments
main "$@"
