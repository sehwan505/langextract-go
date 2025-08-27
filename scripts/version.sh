#!/bin/bash

# LangExtract Version Management Script
# This script helps manage version information and version bumping.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION_FILE="${REPO_ROOT}/VERSION"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_usage() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Version management for LangExtract-Go

COMMANDS:
    show                Show current version
    bump <type>         Bump version (major, minor, patch)
    set <version>       Set specific version
    next <type>         Show what the next version would be

OPTIONS:
    -h, --help          Show this help message
    -d, --dry-run       Show what would be done without making changes
    -f, --file          Read/write version from/to VERSION file

EXAMPLES:
    $0 show             Show current version
    $0 bump patch       Bump patch version (1.0.0 -> 1.0.1)
    $0 bump minor       Bump minor version (1.0.0 -> 1.1.0)
    $0 bump major       Bump major version (1.0.0 -> 2.0.0)
    $0 set v1.2.3       Set version to v1.2.3
    $0 next minor       Show what the next minor version would be

EOF
}

get_current_version() {
    local version=""
    
    # Try to get version from git tags first
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        return 0
    fi
    
    # Try to get version from git tags (latest)
    version=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -n "$version" ]; then
        echo "$version"
        return 0
    fi
    
    # Try to get version from VERSION file
    if [ -f "$VERSION_FILE" ]; then
        version=$(cat "$VERSION_FILE" | tr -d '[:space:]')
        if [ -n "$version" ]; then
            echo "$version"
            return 0
        fi
    fi
    
    # Default version if none found
    echo "v0.0.0"
}

parse_version() {
    local version="$1"
    
    # Remove v prefix if present
    version="${version#v}"
    
    # Split version into parts
    IFS='.' read -r major minor patch <<< "$version"
    
    # Handle pre-release suffixes (e.g., 1.0.0-rc1)
    if [[ "$patch" == *-* ]]; then
        patch="${patch%%-*}"
    fi
    
    # Validate parts are numbers
    if ! [[ "$major" =~ ^[0-9]+$ ]] || ! [[ "$minor" =~ ^[0-9]+$ ]] || ! [[ "$patch" =~ ^[0-9]+$ ]]; then
        log_error "Invalid version format: $1"
        exit 1
    fi
    
    echo "$major $minor $patch"
}

format_version() {
    local major="$1"
    local minor="$2"
    local patch="$3"
    local prefix="${4:-v}"
    
    echo "${prefix}${major}.${minor}.${patch}"
}

bump_version() {
    local current_version="$1"
    local bump_type="$2"
    
    read -r major minor patch <<< "$(parse_version "$current_version")"
    
    case "$bump_type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            log_error "Invalid bump type: $bump_type"
            log_error "Valid types: major, minor, patch"
            exit 1
            ;;
    esac
    
    format_version "$major" "$minor" "$patch"
}

validate_version() {
    local version="$1"
    
    # Check if version follows semantic versioning (with optional v prefix)
    if [[ ! "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        log_error "Invalid version format: $version"
        log_error "Version must follow semantic versioning (e.g., v1.0.0, 1.2.3, v1.0.0-rc1)"
        exit 1
    fi
    
    # Normalize version to include v prefix
    if [[ ! "$version" =~ ^v ]]; then
        version="v$version"
    fi
    
    echo "$version"
}

write_version_file() {
    local version="$1"
    
    if [ "$DRY_RUN" = false ]; then
        echo "$version" > "$VERSION_FILE"
        log_success "Version $version written to $VERSION_FILE"
    else
        log_info "[DRY RUN] Would write version $version to $VERSION_FILE"
    fi
}

show_version() {
    local current_version
    current_version=$(get_current_version)
    
    echo "Current version: $current_version"
    
    # Show additional info
    if git describe --tags --exact-match HEAD 2>/dev/null >/dev/null; then
        log_info "Current commit is tagged with this version"
    else
        local latest_tag
        latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
        local commit_count
        commit_count=$(git rev-list --count HEAD ^$(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD") 2>/dev/null || echo "0")
        
        if [ "$latest_tag" != "none" ] && [ "$commit_count" -gt 0 ]; then
            log_info "Latest tag: $latest_tag"
            log_info "Commits since latest tag: $commit_count"
        fi
    fi
    
    if [ -f "$VERSION_FILE" ]; then
        local file_version
        file_version=$(cat "$VERSION_FILE" | tr -d '[:space:]')
        if [ "$file_version" != "$current_version" ]; then
            log_warning "VERSION file contains: $file_version (differs from git)"
        fi
    fi
}

show_next_version() {
    local bump_type="$1"
    local current_version
    current_version=$(get_current_version)
    local next_version
    next_version=$(bump_version "$current_version" "$bump_type")
    
    echo "Current version: $current_version"
    echo "Next $bump_type version: $next_version"
}

main() {
    local command=""
    local dry_run=false
    local use_file=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -d|--dry-run)
                dry_run=true
                shift
                ;;
            -f|--file)
                use_file=true
                shift
                ;;
            show|bump|set|next)
                if [ -z "$command" ]; then
                    command="$1"
                else
                    log_error "Multiple commands specified"
                    show_usage
                    exit 1
                fi
                shift
                ;;
            -*)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                # This should be command arguments
                break
                ;;
        esac
    done
    
    # Check if command was provided
    if [ -z "$command" ]; then
        log_error "Command is required"
        show_usage
        exit 1
    fi
    
    # Set global variables
    DRY_RUN=$dry_run
    USE_FILE=$use_file
    
    # Execute command
    case "$command" in
        show)
            show_version
            ;;
        bump)
            if [ $# -eq 0 ]; then
                log_error "Bump type is required"
                log_error "Valid types: major, minor, patch"
                exit 1
            fi
            
            local bump_type="$1"
            local current_version
            current_version=$(get_current_version)
            local new_version
            new_version=$(bump_version "$current_version" "$bump_type")
            
            log_info "Bumping $bump_type version: $current_version -> $new_version"
            
            if [ "$USE_FILE" = true ]; then
                write_version_file "$new_version"
            else
                echo "$new_version"
            fi
            ;;
        set)
            if [ $# -eq 0 ]; then
                log_error "Version is required"
                exit 1
            fi
            
            local new_version
            new_version=$(validate_version "$1")
            
            log_info "Setting version to: $new_version"
            
            if [ "$USE_FILE" = true ]; then
                write_version_file "$new_version"
            else
                echo "$new_version"
            fi
            ;;
        next)
            if [ $# -eq 0 ]; then
                log_error "Bump type is required"
                log_error "Valid types: major, minor, patch"
                exit 1
            fi
            
            show_next_version "$1"
            ;;
        *)
            log_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"