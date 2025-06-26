#!/bin/bash

# Development Setup Script for Audit Log Service
# This script sets up the development environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if required tools are installed
check_requirements() {
    print_step "Checking requirements..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | grep -oE '[0-9]+\.[0-9]+')
    REQUIRED_VERSION="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $REQUIRED_VERSION or later is required. Found: $GO_VERSION"
        exit 1
    fi
    
    # Check Docker installation
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not installed. Some features may not work."
    fi
    
    # Check Docker Compose installation
    if ! command -v docker-compose &> /dev/null; then
        print_warning "Docker Compose is not installed. Some features may not work."
    fi
    
    print_success "Requirements check completed"
}

# Install development tools
install_tools() {
    print_step "Installing development tools..."
    
    # Install golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        print_step "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    
    # Install gosec
    if ! command -v gosec &> /dev/null; then
        print_step "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    # Install air for hot reloading
    if ! command -v air &> /dev/null; then
        print_step "Installing air..."
        go install github.com/cosmtrek/air@latest
    fi
    
    # Install migrate tool
    if ! command -v migrate &> /dev/null; then
        print_step "Installing migrate..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    fi
    
    print_success "Development tools installed"
}

# Setup configuration
setup_config() {
    print_step "Setting up configuration..."
    
    if [ ! -f "configs/config.yaml" ]; then
        print_step "Copying example configuration..."
        cp configs/config.example.yaml configs/config.yaml
        print_success "Configuration file created at configs/config.yaml"
        print_warning "Please review and modify the configuration as needed"
    else
        print_warning "Configuration file already exists"
    fi
}

# Download dependencies
download_dependencies() {
    print_step "Downloading Go dependencies..."
    go mod download
    go mod tidy
    print_success "Dependencies downloaded"
}

# Setup database (if Docker is available)
setup_database() {
    if command -v docker &> /dev/null; then
        print_step "Setting up PostgreSQL database..."
        
        # Check if PostgreSQL container is already running
        if docker ps | grep -q "audit-postgres"; then
            print_warning "PostgreSQL container is already running"
        else
            print_step "Starting PostgreSQL container..."
            docker run -d \
                --name audit-postgres \
                -e POSTGRES_USER=postgres \
                -e POSTGRES_PASSWORD=password \
                -e POSTGRES_DB=audit_logs \
                -p 5432:5432 \
                postgres:15-alpine
            
            # Wait for PostgreSQL to be ready
            print_step "Waiting for PostgreSQL to be ready..."
            sleep 10
            
            # Check if PostgreSQL is ready
            for i in {1..30}; do
                if docker exec audit-postgres pg_isready -U postgres; then
                    break
                fi
                sleep 1
            done
        fi
        
        print_success "PostgreSQL database is ready"
    else
        print_warning "Docker not available. Please set up PostgreSQL manually"
    fi
}

# Run database migrations
run_migrations() {
    print_step "Running database migrations..."
    
    # Set database environment variables
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USER=postgres
    export DB_PASSWORD=password
    export DB_NAME=audit_logs
    
    # Run migrations
    ./scripts/migrate.sh up
    
    print_success "Database migrations completed"
}

# Create logs directory
create_directories() {
    print_step "Creating necessary directories..."
    mkdir -p logs
    mkdir -p tmp
    print_success "Directories created"
}

# Run tests
run_tests() {
    print_step "Running tests..."
    go test ./... -v
    print_success "Tests completed"
}

# Main setup function
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                  Audit Log Service Setup                    ║"
    echo "║                                                              ║"
    echo "║  This script will set up your development environment       ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    check_requirements
    install_tools
    setup_config
    download_dependencies
    create_directories
    setup_database
    run_migrations
    run_tests
    
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    Setup Complete!                          ║"
    echo "║                                                              ║"
    echo "║  Your development environment is ready.                     ║"
    echo "║                                                              ║"
    echo "║  Quick start commands:                                       ║"
    echo "║    make dev      - Start development server with hot reload ║"
    echo "║    make test     - Run tests                                 ║"
    echo "║    make lint     - Run linters                               ║"
    echo "║    make help     - Show all available commands               ║"
    echo "║                                                              ║"
    echo "║  Access the service at: http://localhost:9025                ║"
    echo "║  Health check: http://localhost:9025/health                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Run the main function
main "$@"
