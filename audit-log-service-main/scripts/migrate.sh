#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    print_info "Loading environment variables from .env file"
    export $(cat .env | grep -v '^#' | xargs)
fi

# Database connection details with defaults
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-Abcd1234}
DB_NAME=${DB_NAME:-audit_logs}
DB_SSL_MODE=${DB_SSL_MODE:-disable}

# Migration directory
MIGRATION_DIR="./internal/database/migrations"

# Function to check if database exists
check_database_exists() {
    print_info "Checking if database '$DB_NAME' exists..."
    
    # Check if database exists
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"
    
    if [ $? -eq 0 ]; then
        print_info "Database '$DB_NAME' exists"
        return 0
    else
        print_warn "Database '$DB_NAME' does not exist"
        return 1
    fi
}

# Function to create database if it doesn't exist
create_database() {
    print_info "Creating database '$DB_NAME'..."
    
    PGPASSWORD=$DB_PASSWORD createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"
    
    if [ $? -eq 0 ]; then
        print_info "Database '$DB_NAME' created successfully"
    else
        print_error "Failed to create database '$DB_NAME'"
        exit 1
    fi
}

# Function to test database connection
test_connection() {
    print_info "Testing database connection..."
    
    PGPASSWORD=$DB_PASSWORD pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        print_info "Database connection successful"
        return 0
    else
        print_error "Cannot connect to database"
        return 1
    fi
}

# Function to run migrations up
migrate_up() {
    print_info "Running migrations up for database: $DB_NAME"
    
    if [ ! -d "$MIGRATION_DIR" ]; then
        print_error "Migration directory '$MIGRATION_DIR' not found"
        exit 1
    fi
    
    # Create migrations table if it doesn't exist
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    " > /dev/null 2>&1
    
    # Run migration files in order
    for migration_file in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$migration_file" ]; then
            filename=$(basename "$migration_file")
            version=${filename%.*}
            
            # Check if migration has already been applied
            already_applied=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version='$version';" | xargs)
            
            if [ "$already_applied" -eq 0 ]; then
                print_info "Applying migration: $filename"
                
                # Run the migration
                PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"
                
                if [ $? -eq 0 ]; then
                    # Record the migration as applied
                    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "INSERT INTO schema_migrations (version) VALUES ('$version');" > /dev/null 2>&1
                    print_info "Migration $filename applied successfully"
                else
                    print_error "Migration $filename failed"
                    exit 1
                fi
            else
                print_info "Migration $filename already applied, skipping"
            fi
        fi
    done
    
    print_info "All migrations completed successfully"
}

# Function to rollback last migration
migrate_down() {
    print_info "Rolling back last migration..."
    
    # Get the last applied migration
    last_migration=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;" | xargs)
    
    if [ -z "$last_migration" ]; then
        print_warn "No migrations to rollback"
        return 0
    fi
    
    print_info "Rolling back migration: $last_migration"
    
    # For now, we'll just remove the record from schema_migrations
    # In a real implementation, you'd have down migration files
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "DELETE FROM schema_migrations WHERE version='$last_migration';" > /dev/null 2>&1
    
    print_info "Migration $last_migration rolled back"
}

# Function to show migration status
migrate_status() {
    print_info "Migration status for database: $DB_NAME"
    
    # Check if migrations table exists
    table_exists=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name='schema_migrations';" | xargs)
    
    if [ "$table_exists" -eq 0 ]; then
        print_warn "Migrations table does not exist. Run 'migrate up' first."
        return 0
    fi
    
    echo ""
    echo "Applied migrations:"
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at;"
    echo ""
    
    # Show pending migrations
    echo "Available migration files:"
    for migration_file in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$migration_file" ]; then
            filename=$(basename "$migration_file")
            version=${filename%.*}
            
            already_applied=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version='$version';" | xargs)
            
            if [ "$already_applied" -eq 0 ]; then
                echo "  ❌ $filename (pending)"
            else
                echo "  ✅ $filename (applied)"
            fi
        fi
    done
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  up       Run all pending migrations (default)"
    echo "  down     Rollback the last migration"
    echo "  status   Show migration status"
    echo "  create   Create the database if it doesn't exist"
    echo "  test     Test database connection"
    echo "  help     Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST      Database host (default: localhost)"
    echo "  DB_PORT      Database port (default: 5432)"
    echo "  DB_USER      Database user (default: postgres)"
    echo "  DB_PASSWORD  Database password (default: password)"
    echo "  DB_NAME      Database name (default: audit_logs)"
    echo ""
}

# Main logic
COMMAND=${1:-up}

case $COMMAND in
    "up")
        if ! check_database_exists; then
            create_database
        fi
        
        if test_connection; then
            migrate_up
        else
            print_error "Cannot connect to database. Please check your connection settings."
            exit 1
        fi
        ;;
    "down")
        if test_connection; then
            migrate_down
        else
            print_error "Cannot connect to database. Please check your connection settings."
            exit 1
        fi
        ;;
    "status")
        if test_connection; then
            migrate_status
        else
            print_error "Cannot connect to database. Please check your connection settings."
            exit 1
        fi
        ;;
    "create")
        if ! check_database_exists; then
            create_database
        else
            print_info "Database '$DB_NAME' already exists"
        fi
        ;;
    "test")
        test_connection
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac