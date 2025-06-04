#!/bin/bash

# Database Migration Runner Script for Instagram-like Social Media Platform
# This script runs all migrations in the correct order

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-fowergram}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-password}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./migrations}"

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

# Function to check if PostgreSQL is available
check_postgres() {
    print_status "Checking PostgreSQL connection..."
    
    if ! command -v psql &> /dev/null; then
        print_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi
    
    export PGPASSWORD="$DB_PASSWORD"
    
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT 1;" &> /dev/null; then
        print_error "Cannot connect to PostgreSQL server at $DB_HOST:$DB_PORT"
        print_error "Please check your database connection settings."
        exit 1
    fi
    
    print_success "PostgreSQL connection successful"
}

# Function to create database if it doesn't exist
create_database() {
    print_status "Checking if database '$DB_NAME' exists..."
    
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        print_success "Database '$DB_NAME' already exists"
    else
        print_status "Creating database '$DB_NAME'..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"
        print_success "Database '$DB_NAME' created successfully"
    fi
}

# Function to create migrations table
create_migrations_table() {
    print_status "Creating migrations tracking table..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << 'EOF'
CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL UNIQUE,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64)
);
EOF
    
    print_success "Migrations table ready"
}

# Function to check if migration was already executed
is_migration_executed() {
    local filename="$1"
    local count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE filename = '$filename';")
    [ "$count" -gt 0 ]
}

# Function to calculate file checksum
calculate_checksum() {
    local file="$1"
    if command -v sha256sum &> /dev/null; then
        sha256sum "$file" | cut -d' ' -f1
    elif command -v shasum &> /dev/null; then
        shasum -a 256 "$file" | cut -d' ' -f1
    else
        # Fallback to a simple hash
        wc -c < "$file"
    fi
}

# Function to record migration execution
record_migration() {
    local filename="$1"
    local checksum="$2"
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        INSERT INTO schema_migrations (filename, checksum) 
        VALUES ('$filename', '$checksum')
        ON CONFLICT (filename) DO UPDATE SET 
            executed_at = CURRENT_TIMESTAMP,
            checksum = EXCLUDED.checksum;
    " > /dev/null
}

# Function to run a single migration
run_migration() {
    local migration_file="$1"
    local filename=$(basename "$migration_file")
    
    if is_migration_executed "$filename"; then
        print_warning "Migration '$filename' already executed, skipping..."
        return 0
    fi
    
    print_status "Running migration: $filename"
    
    local checksum=$(calculate_checksum "$migration_file")
    
    # Run the migration
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"; then
        record_migration "$filename" "$checksum"
        print_success "Migration '$filename' completed successfully"
        return 0
    else
        print_error "Migration '$filename' failed!"
        return 1
    fi
}

# Function to rollback a migration
rollback_migration() {
    local migration_file="$1"
    local filename=$(basename "$migration_file" .sql)
    local down_file="${migration_file%.*}.down.sql"
    
    if [ ! -f "$down_file" ]; then
        print_error "Rollback file not found: $down_file"
        return 1
    fi
    
    print_status "Rolling back migration: $filename"
    
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$down_file"; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "DELETE FROM schema_migrations WHERE filename = '$(basename "$migration_file")';" > /dev/null
        print_success "Rollback of '$filename' completed successfully"
        return 0
    else
        print_error "Rollback of '$filename' failed!"
        return 1
    fi
}

# Function to run all migrations
run_all_migrations() {
    print_status "Starting migration process..."
    
    # Define migration order
    local migrations=(
        "001_initial_schema.sql"
        "002_jwt_auth_schema.sql"
        "000003_create_verification_tables.up.sql"
        "004_instagram_optimization.sql"
        "005_instagram_advanced_features.sql"
    )
    
    local success_count=0
    local total_count=${#migrations[@]}
    
    for migration in "${migrations[@]}"; do
        local migration_path="$MIGRATIONS_DIR/$migration"
        
        if [ ! -f "$migration_path" ]; then
            print_warning "Migration file not found: $migration_path"
            continue
        fi
        
        if run_migration "$migration_path"; then
            ((success_count++))
        else
            print_error "Migration process stopped due to failure"
            exit 1
        fi
    done
    
    print_success "Migration process completed: $success_count/$total_count migrations executed"
}

# Function to show migration status
show_status() {
    print_status "Migration Status:"
    echo
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            filename,
            executed_at,
            checksum
        FROM schema_migrations 
        ORDER BY executed_at;
    "
}

# Function to rollback last migration
rollback_last() {
    local last_migration=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT filename FROM schema_migrations 
        ORDER BY executed_at DESC 
        LIMIT 1;
    " | tr -d ' ')
    
    if [ -z "$last_migration" ]; then
        print_warning "No migrations to rollback"
        return 0
    fi
    
    local migration_path="$MIGRATIONS_DIR/$last_migration"
    rollback_migration "$migration_path"
}

# Function to reset database
reset_database() {
    print_warning "This will DROP and recreate the database. Are you sure? (y/N)"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_status "Dropping database '$DB_NAME'..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
        
        print_status "Creating fresh database '$DB_NAME'..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"
        
        create_migrations_table
        run_all_migrations
        
        print_success "Database reset completed"
    else
        print_status "Database reset cancelled"
    fi
}

# Function to show help
show_help() {
    echo "Instagram Database Migration Script"
    echo
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo
    echo "Commands:"
    echo "  migrate         Run all pending migrations (default)"
    echo "  status          Show migration status"
    echo "  rollback        Rollback the last migration"
    echo "  reset           Drop and recreate database with all migrations"
    echo "  help            Show this help message"
    echo
    echo "Environment Variables:"
    echo "  DB_HOST         Database host (default: localhost)"
    echo "  DB_PORT         Database port (default: 5432)"
    echo "  DB_NAME         Database name (default: fowergram)"
    echo "  DB_USER         Database user (default: postgres)"
    echo "  DB_PASSWORD     Database password (default: password)"
    echo "  MIGRATIONS_DIR  Migrations directory (default: ./migrations)"
    echo
    echo "Examples:"
    echo "  $0                                    # Run all migrations"
    echo "  $0 status                             # Show migration status"
    echo "  $0 rollback                           # Rollback last migration"
    echo "  DB_NAME=mydb $0 migrate               # Run migrations on specific database"
}

# Main script logic
main() {
    local command="${1:-migrate}"
    
    case "$command" in
        "migrate")
            check_postgres
            create_database
            create_migrations_table
            run_all_migrations
            ;;
        "status")
            check_postgres
            show_status
            ;;
        "rollback")
            check_postgres
            rollback_last
            ;;
        "reset")
            check_postgres
            reset_database
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 