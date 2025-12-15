#!/bin/bash

# Test runner script untuk UAS project
# Script ini menjalankan berbagai jenis test dengan coverage reporting

set -e

echo "🧪 Menjalankan UAS Project Tests"
echo "================================"

# Colors untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function untuk print colored output
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

# Check jika go terinstall
if ! command -v go &> /dev/null; then
    print_error "Go tidak terinstall atau tidak ada di PATH"
    exit 1
fi

# Navigate ke project root
cd "$(dirname "$0")/.."

print_status "Current directory: $(pwd)"

# Clean previous test artifacts
print_status "Membersihkan artifacts test sebelumnya..."
rm -f coverage.out coverage.html

# Download dependencies
print_status "Mendownload dependencies..."
go mod download
go mod tidy

# Run linting (jika golangci-lint tersedia)
if command -v golangci-lint &> /dev/null; then
    print_status "Menjalankan linter..."
    golangci-lint run ./... || print_warning "Ditemukan linting issues"
else
    print_warning "golangci-lint tidak ditemukan, skip linting"
fi

# Run unit tests
print_status "Menjalankan unit tests..."
go test -v -race -coverprofile=coverage.out ./test/... || {
    print_error "Unit tests gagal"
    exit 1
}

print_success "Unit tests berhasil!"

# Generate coverage report
if [ -f coverage.out ]; then
    print_status "Membuat coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    
    # Show coverage summary
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_success "Total coverage: $COVERAGE"
    
    if command -v open &> /dev/null; then
        print_status "Membuka coverage report di browser..."
        open coverage.html
    elif command -v xdg-open &> /dev/null; then
        print_status "Membuka coverage report di browser..."
        xdg-open coverage.html
    else
        print_status "Coverage report disimpan ke coverage.html"
    fi
else
    print_warning "Tidak ada coverage report yang dibuat"
fi

# Run benchmarks (jika ada)
print_status "Menjalankan benchmarks..."
go test -bench=. -benchmem ./test/... || print_warning "Tidak ada benchmarks ditemukan"

# Test build
print_status "Testing build..."
go build -o /tmp/uas_test ./main.go || {
    print_error "Build gagal"
    exit 1
}

rm -f /tmp/uas_test
print_success "Build test berhasil!"

echo ""
print_success "Semua tests selesai dengan sukses! 🎉"
echo ""
echo "📊 Test Summary:"
echo "  - Unit tests: ✅ Passed"
echo "  - Build test: ✅ Passed"
if [ -f coverage.out ]; then
    echo "  - Coverage: $COVERAGE"
fi
echo ""
echo "📁 Generated files:"
echo "  - coverage.out (coverage data)"
echo "  - coverage.html (coverage report)"
echo ""
echo "💡 Untuk menjalankan test individual:"
echo "  go test -v ./test/utils_test.go"
echo "  go test -v ./test/service_test.go"
echo "  go test -v ./test/repository_test.go"
echo "  go test -v ./test/middleware_test.go"
echo ""