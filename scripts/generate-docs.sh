#!/bin/bash

# LangExtract-Go Documentation Generation Script
# Generates comprehensive API documentation from Go code

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCS_DIR="${REPO_ROOT}/docs"
API_DOCS_DIR="${DOCS_DIR}/api"
TEMP_DIR="${REPO_ROOT}/tmp/docs"

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
Usage: $0 [OPTIONS]

Generate comprehensive API documentation for LangExtract-Go

OPTIONS:
    -h, --help          Show this help message
    -f, --format FORMAT Output format: html, markdown, json (default: html)
    -o, --output DIR    Output directory (default: docs/api)
    -c, --clean         Clean existing docs before generating
    -s, --serve         Start local server to view docs after generation
    -p, --port PORT     Port for local server (default: 6060)

EXAMPLES:
    $0                  Generate HTML documentation
    $0 -f markdown      Generate Markdown documentation
    $0 -c -s           Clean, generate, and serve documentation
    $0 --output /tmp   Generate docs in /tmp directory

EOF
}

# Check dependencies
check_dependencies() {
    log_info "Checking documentation generation dependencies..."
    
    # Check if godoc is available
    if ! command -v godoc &> /dev/null; then
        log_info "Installing godoc..."
        go install golang.org/x/tools/cmd/godoc@latest
    fi
    
    # Check if go doc is available (should be built-in)
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Clean existing documentation
clean_docs() {
    if [ -d "$API_DOCS_DIR" ]; then
        log_info "Cleaning existing documentation..."
        rm -rf "$API_DOCS_DIR"
        log_success "Existing documentation cleaned"
    fi
}

# Generate package list
generate_package_list() {
    # Get all Go packages (no logging to avoid contaminating output)
    local packages=()
    while IFS= read -r -d '' file; do
        local dir=$(dirname "$file")
        # Make path relative to repo root (macOS compatible)
        local rel_dir="${dir#$REPO_ROOT/}"
        if [[ "$rel_dir" == pkg/* ]] || [[ "$rel_dir" == cmd/* ]] || [[ "$rel_dir" == internal/* ]]; then
            packages+=("github.com/sehwan505/langextract-go/$rel_dir")
        fi
    done < <(find "$REPO_ROOT" -name "*.go" -not -path "*/test/*" -not -path "*/.*" -not -name "*_test.go" -print0)
    
    # Remove duplicates and sort
    if [ ${#packages[@]} -gt 0 ]; then
        printf '%s\n' "${packages[@]}" | sort -u
    fi
}

# Generate HTML documentation
generate_html_docs() {
    log_info "Generating HTML documentation..."
    
    mkdir -p "$API_DOCS_DIR/html"
    
    # Generate package documentation  
    log_info "Discovering packages..."
    local packages_list
    packages_list=$(generate_package_list)
    
    if [ -z "$packages_list" ]; then
        log_warning "No packages found for documentation generation"
        return 0
    fi
    
    while IFS= read -r package; do
        [ -z "$package" ] && continue
        
        local pkg_name=$(basename "$package")
        local pkg_path=$(echo "$package" | sed 's|github.com/sehwan505/langextract-go/||')
        
        log_info "Generating docs for $pkg_name..."
        
        # Create package directory
        local pkg_dir="$API_DOCS_DIR/html/$pkg_name"
        mkdir -p "$pkg_dir"
        
        # Generate package documentation using go doc
        go doc -all "$package" > "$pkg_dir/index.txt" 2>/dev/null || {
            log_warning "Could not generate docs for $package"
            continue
        }
        
        # Convert to HTML format
        cat > "$pkg_dir/index.html" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>$pkg_name - LangExtract-Go API Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #fafafa;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            border-bottom: 2px solid #e9ecef;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .package-name {
            color: #2c3e50;
            font-size: 2em;
            margin: 0;
            font-weight: 600;
        }
        .package-path {
            color: #7f8c8d;
            font-family: 'Monaco', 'Menlo', monospace;
            margin-top: 5px;
        }
        .content {
            white-space: pre-wrap;
            font-family: 'Monaco', 'Menlo', monospace;
            background: #f8f9fa;
            padding: 20px;
            border-radius: 4px;
            overflow-x: auto;
        }
        .nav {
            margin-bottom: 20px;
        }
        .nav a {
            color: #3498db;
            text-decoration: none;
            margin-right: 15px;
        }
        .nav a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="nav">
            <a href="../index.html">← Back to API Index</a>
        </div>
        <div class="header">
            <h1 class="package-name">$pkg_name</h1>
            <div class="package-path">$package</div>
        </div>
        <div class="content">$(cat "$pkg_dir/index.txt" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')</div>
    </div>
</body>
</html>
EOF
    done <<< "$packages_list"
    
    # Generate main index  
    generate_html_index "$packages_list"
    
    log_success "HTML documentation generated in $API_DOCS_DIR/html"
}

# Generate HTML index page
generate_html_index() {
    local packages_list="$1"
    
    if [ -z "$packages_list" ]; then
        log_warning "No packages provided for index generation"
        return 0
    fi
    
    log_info "Generating API documentation index..."
    
    cat > "$API_DOCS_DIR/html/index.html" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LangExtract-Go API Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #fafafa;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            border-bottom: 2px solid #e9ecef;
            padding-bottom: 30px;
            margin-bottom: 40px;
        }
        .title {
            color: #2c3e50;
            font-size: 2.5em;
            margin: 0 0 10px 0;
            font-weight: 600;
        }
        .subtitle {
            color: #7f8c8d;
            font-size: 1.1em;
            margin: 0;
        }
        .package-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .package-card {
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            background: #f8f9fa;
            transition: box-shadow 0.3s ease;
        }
        .package-card:hover {
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .package-card h3 {
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .package-card a {
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }
        .package-card a:hover {
            text-decoration: underline;
        }
        .package-path {
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
            color: #7f8c8d;
            background: white;
            padding: 5px 8px;
            border-radius: 4px;
            margin-top: 10px;
        }
        .stats {
            display: flex;
            justify-content: space-around;
            margin: 30px 0;
            padding: 20px;
            background: #e3f2fd;
            border-radius: 8px;
        }
        .stat {
            text-align: center;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #1976d2;
        }
        .stat-label {
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">LangExtract-Go</h1>
            <p class="subtitle">API Documentation</p>
        </div>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-number">$(echo "$packages_list" | wc -l)</div>
                <div class="stat-label">Packages</div>
            </div>
            <div class="stat">
                <div class="stat-number">$(find "$REPO_ROOT" -name "*.go" -not -path "*/test/*" -not -name "*_test.go" | wc -l)</div>
                <div class="stat-label">Source Files</div>
            </div>
            <div class="stat">
                <div class="stat-number">$(date '+%Y-%m-%d')</div>
                <div class="stat-label">Generated</div>
            </div>
        </div>
        
        <div class="package-grid">
EOF

    # Add package cards
    while IFS= read -r package; do
        [ -z "$package" ] && continue
        local pkg_name=$(basename "$package")
        local pkg_path=$(echo "$package" | sed 's|github.com/sehwan505/langextract-go/||')
        local description=""
        
        # Try to get package description
        if [ -f "$REPO_ROOT/$pkg_path/doc.go" ]; then
            description=$(head -n 10 "$REPO_ROOT/$pkg_path/doc.go" | grep -E '^// [A-Z]' | head -n 1 | sed 's|^// ||' || echo "")
        fi
        
        if [ -z "$description" ]; then
            case "$pkg_name" in
                "langextract") description="Main public API for text extraction" ;;
                "document") description="Document structures and annotations" ;;
                "extraction") description="Extraction types and schema validation" ;;
                "providers") description="LLM provider implementations" ;;
                "types") description="Core data types and intervals" ;;
                "engine") description="Internal extraction engine" ;;
                "alignment") description="Text alignment algorithms" ;;
                "chunking") description="Document chunking strategies" ;;
                "prompt") description="Prompt engineering and templates" ;;
                "visualization") description="Output formatting and visualization" ;;
                *) description="Go package for LangExtract functionality" ;;
            esac
        fi
        
        cat >> "$API_DOCS_DIR/html/index.html" << EOF
            <div class="package-card">
                <h3><a href="$pkg_name/index.html">$pkg_name</a></h3>
                <p>$description</p>
                <div class="package-path">$package</div>
            </div>
EOF
    done <<< "$packages_list"
    
    cat >> "$API_DOCS_DIR/html/index.html" << EOF
        </div>
    </div>
</body>
</html>
EOF
    
    log_success "API documentation index generated"
}

# Generate Markdown documentation
generate_markdown_docs() {
    log_info "Generating Markdown documentation..."
    
    mkdir -p "$API_DOCS_DIR/markdown"
    
    # Generate package documentation  
    local packages_list
    packages_list=$(generate_package_list)
    
    if [ -z "$packages_list" ]; then
        log_warning "No packages found for documentation generation"
        return 0
    fi
    
    # Generate main README
    cat > "$API_DOCS_DIR/markdown/README.md" << EOF
# LangExtract-Go API Documentation

Generated on: $(date)

## Packages

EOF

    while IFS= read -r package; do
        [ -z "$package" ] && continue
        local pkg_name=$(basename "$package")
        local pkg_path=$(echo "$package" | sed 's|github.com/sehwan505/langextract-go/||')
        
        log_info "Generating Markdown docs for $pkg_name..."
        
        # Generate package documentation
        echo "- [$pkg_name](./$pkg_name.md) - $package" >> "$API_DOCS_DIR/markdown/README.md"
        
        # Create individual package markdown
        cat > "$API_DOCS_DIR/markdown/$pkg_name.md" << EOF
# $pkg_name

Package: \`$package\`

\`\`\`go
EOF
        go doc -all "$package" >> "$API_DOCS_DIR/markdown/$pkg_name.md" 2>/dev/null || {
            echo "// Documentation not available" >> "$API_DOCS_DIR/markdown/$pkg_name.md"
        }
        echo '```' >> "$API_DOCS_DIR/markdown/$pkg_name.md"
    done <<< "$packages_list"
    
    log_success "Markdown documentation generated in $API_DOCS_DIR/markdown"
}

# Start local documentation server
serve_docs() {
    local port=${1:-6060}
    local format=${2:-html}
    
    if [ "$format" = "html" ] && [ -d "$API_DOCS_DIR/html" ]; then
        log_info "Starting documentation server on port $port..."
        log_info "Documentation available at: http://localhost:$port"
        log_info "Press Ctrl+C to stop the server"
        
        cd "$API_DOCS_DIR/html"
        if command -v python3 &> /dev/null; then
            python3 -m http.server "$port"
        elif command -v python &> /dev/null; then
            python -m SimpleHTTPServer "$port"
        else
            log_error "Python is required to serve documentation"
            exit 1
        fi
    else
        log_error "HTML documentation not found. Generate it first with: $0 -f html"
        exit 1
    fi
}

# Main function
main() {
    local format="html"
    local output_dir="$API_DOCS_DIR"
    local clean_first=false
    local serve_after=false
    local port=6060
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -f|--format)
                format="$2"
                shift 2
                ;;
            -o|--output)
                output_dir="$2"
                API_DOCS_DIR="$2"
                shift 2
                ;;
            -c|--clean)
                clean_first=true
                shift
                ;;
            -s|--serve)
                serve_after=true
                shift
                ;;
            -p|--port)
                port="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Validate format
    if [[ "$format" != "html" && "$format" != "markdown" && "$format" != "json" ]]; then
        log_error "Invalid format: $format. Supported formats: html, markdown, json"
        exit 1
    fi
    
    # Ensure we're in the right directory
    cd "$REPO_ROOT"
    
    # Check dependencies
    check_dependencies
    
    # Clean if requested
    if [ "$clean_first" = true ]; then
        clean_docs
    fi
    
    # Create output directory
    mkdir -p "$output_dir"
    
    # Generate documentation based on format
    case "$format" in
        html)
            generate_html_docs
            ;;
        markdown)
            generate_markdown_docs
            ;;
        json)
            log_error "JSON format not yet implemented"
            exit 1
            ;;
    esac
    
    # Serve if requested
    if [ "$serve_after" = true ]; then
        serve_docs "$port" "$format"
    fi
    
    log_success "Documentation generation completed!"
    log_info "Output directory: $output_dir"
}

# Run main function with all arguments
main "$@"