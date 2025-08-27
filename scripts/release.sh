#!/bin/bash

# LangExtract Release Script
# This script automates the release process including versioning, changelog generation, and release builds.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHANGELOG_FILE="${REPO_ROOT}/CHANGELOG.md"
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
Usage: $0 [OPTIONS] <version>

Release script for LangExtract-Go

OPTIONS:
    -h, --help          Show this help message
    -d, --dry-run       Perform a dry run without making changes
    -p, --pre-release   Mark as pre-release
    -f, --force         Force release even if there are uncommitted changes
    -t, --tag-only      Only create and push the tag (no builds)

EXAMPLES:
    $0 v1.0.0           Create version v1.0.0 release
    $0 -d v1.0.1        Dry run for version v1.0.1
    $0 -p v1.0.0-rc1    Create pre-release v1.0.0-rc1
    $0 -t v1.0.2        Only create and push tag v1.0.2

EOF
}

check_dependencies() {
    local missing_deps=()
    
    command -v git >/dev/null 2>&1 || missing_deps+=("git")
    command -v make >/dev/null 2>&1 || missing_deps+=("make")
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again."
        exit 1
    fi
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

check_git_status() {
    if [ "$FORCE" = false ] && ! git diff-index --quiet HEAD --; then
        log_error "There are uncommitted changes in the repository"
        log_error "Please commit or stash your changes before creating a release"
        log_error "Use --force to override this check"
        exit 1
    fi
    
    # Check if we're on main branch
    local current_branch
    current_branch=$(git branch --show-current)
    if [ "$current_branch" != "main" ] && [ "$current_branch" != "master" ]; then
        log_warning "Not on main/master branch (current: $current_branch)"
        if [ "$FORCE" = false ]; then
            log_error "Use --force to release from non-main branch"
            exit 1
        fi
    fi
}

check_existing_tag() {
    local version="$1"
    
    if git tag -l | grep -q "^$version$"; then
        log_error "Tag $version already exists"
        log_error "Please choose a different version or delete the existing tag"
        exit 1
    fi
}

generate_changelog() {
    local version="$1"
    local previous_tag
    
    # Get the previous tag
    previous_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    log_info "Generating changelog for $version..."
    
    # Create changelog entry
    local changelog_entry
    changelog_entry="## [$version] - $(date -u +%Y-%m-%d)

### Changes"
    
    if [ -n "$previous_tag" ]; then
        log_info "Comparing with previous tag: $previous_tag"
        changelog_entry="$changelog_entry
$(git log "$previous_tag..HEAD" --pretty=format:"- %s" --reverse)"
    else
        log_info "No previous tag found, including all commits"
        changelog_entry="$changelog_entry
$(git log --pretty=format:"- %s" --reverse)"
    fi
    
    # Update CHANGELOG.md
    if [ -f "$CHANGELOG_FILE" ]; then
        # Insert new entry after the header
        local temp_file
        temp_file=$(mktemp)
        {
            head -n 2 "$CHANGELOG_FILE" 2>/dev/null || echo -e "# Changelog\n"
            echo ""
            echo "$changelog_entry"
            echo ""
            if [ -s "$CHANGELOG_FILE" ]; then
                tail -n +3 "$CHANGELOG_FILE" 2>/dev/null || true
            fi
        } > "$temp_file"
        mv "$temp_file" "$CHANGELOG_FILE"
    else
        # Create new CHANGELOG.md
        {
            echo "# Changelog"
            echo ""
            echo "$changelog_entry"
            echo ""
        } > "$CHANGELOG_FILE"
    fi
    
    # Save version to VERSION file
    echo "$version" > "$VERSION_FILE"
    
    log_success "Changelog updated"
}

build_release() {
    local version="$1"
    
    log_info "Building release for $version..."
    
    cd "$REPO_ROOT"
    
    # Build release binaries
    if [ "$DRY_RUN" = false ]; then
        make clean
        VERSION="$version" make release
        VERSION="$version" make dist
        log_success "Release builds completed"
    else
        log_info "[DRY RUN] Would build release binaries"
    fi
}

create_git_tag() {
    local version="$1"
    
    log_info "Creating git tag $version..."
    
    if [ "$DRY_RUN" = false ]; then
        # Add changed files
        git add "$CHANGELOG_FILE" "$VERSION_FILE" 2>/dev/null || true
        
        # Commit changelog and version file
        if ! git diff --cached --quiet; then
            git commit -m "chore: release $version

- Update CHANGELOG.md
- Update VERSION file

🤖 Generated with [Claude Code](https://claude.ai/code)"
        fi
        
        # Create tag
        local tag_message="Release $version

$(head -n 20 "$CHANGELOG_FILE" | tail -n +4)"
        
        git tag -a "$version" -m "$tag_message"
        
        log_success "Git tag $version created"
        
        # Push tag
        log_info "Pushing tag to remote..."
        git push origin "$version"
        git push origin HEAD
        
        log_success "Tag pushed to remote"
    else
        log_info "[DRY RUN] Would create git tag $version"
        log_info "[DRY RUN] Would push tag to remote"
    fi
}

main() {
    local version=""
    local dry_run=false
    local pre_release=false
    local force=false
    local tag_only=false
    
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
            -p|--pre-release)
                pre_release=true
                shift
                ;;
            -f|--force)
                force=true
                shift
                ;;
            -t|--tag-only)
                tag_only=true
                shift
                ;;
            -*)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [ -z "$version" ]; then
                    version="$1"
                else
                    log_error "Multiple versions specified"
                    show_usage
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    # Check if version was provided
    if [ -z "$version" ]; then
        log_error "Version is required"
        show_usage
        exit 1
    fi
    
    # Set global variables
    DRY_RUN=$dry_run
    PRE_RELEASE=$pre_release
    FORCE=$force
    TAG_ONLY=$tag_only
    
    # Validate version
    version=$(validate_version "$version")
    
    log_info "Starting release process for version: $version"
    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN MODE - No changes will be made"
    fi
    
    # Check dependencies and environment
    check_dependencies
    check_git_status
    check_existing_tag "$version"
    
    # Generate changelog and update version
    generate_changelog "$version"
    
    # Build release if not tag-only
    if [ "$TAG_ONLY" = false ]; then
        build_release "$version"
    fi
    
    # Create and push git tag
    create_git_tag "$version"
    
    log_success "Release $version completed successfully!"
    
    if [ "$DRY_RUN" = false ]; then
        echo ""
        log_info "Next steps:"
        log_info "1. Check the GitHub release page"
        log_info "2. Upload release artifacts if needed"
        log_info "3. Announce the release"
        echo ""
    fi
}

# Run main function with all arguments
main "$@"