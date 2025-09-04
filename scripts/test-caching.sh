#!/bin/bash

# Asset Caching Test Script
# This script demonstrates how the caching implementation works by making HTTP requests
# and showing the cache headers returned by the server.

set -e

# Configuration
SERVER_URL="http://localhost:9020"
SERVER_PATH_PREFIX="/pg-press"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print colored output
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to test cache headers for a specific asset
test_cache_headers() {
    local asset_path="$1"
    local expected_cache_type="$2"
    local url="${SERVER_URL}${SERVER_PATH_PREFIX}${asset_path}"

    print_header "Testing: ${asset_path}"
    echo "URL: ${url}"

    # First request - should get full response with cache headers
    echo -e "\n${YELLOW}First request (should receive full response):${NC}"
    response=$(curl -s -I "${url}" 2>/dev/null || echo "ERROR")

    if [[ "$response" == "ERROR" ]]; then
        print_error "Failed to connect to server at ${url}"
        print_warning "Make sure the server is running with: make dev"
        return 1
    fi

    # Check status code
    status_code=$(echo "$response" | head -n1 | cut -d' ' -f2)
    if [[ "$status_code" == "200" ]]; then
        print_success "Status: ${status_code} OK"
    else
        print_error "Status: ${status_code} (expected 200)"
    fi

    # Check cache headers
    cache_control=$(echo "$response" | grep -i "cache-control" | cut -d' ' -f2- | tr -d '\r')
    expires=$(echo "$response" | grep -i "expires" | cut -d' ' -f2- | tr -d '\r')
    etag=$(echo "$response" | grep -i "etag" | cut -d' ' -f2- | tr -d '\r')
    last_modified=$(echo "$response" | grep -i "last-modified" | cut -d' ' -f2- | tr -d '\r')
    vary=$(echo "$response" | grep -i "vary" | cut -d' ' -f2- | tr -d '\r')

    echo "Cache-Control: ${cache_control:-"Not set"}"
    echo "Expires: ${expires:-"Not set"}"
    echo "ETag: ${etag:-"Not set"}"
    echo "Last-Modified: ${last_modified:-"Not set"}"
    echo "Vary: ${vary:-"Not set"}"

    # Validate cache headers based on expected type
    case "$expected_cache_type" in
        "long-term")
            if [[ "$cache_control" == *"max-age=31536000"* && "$cache_control" == *"immutable"* ]]; then
                print_success "Long-term caching headers are correct"
            else
                print_error "Expected long-term caching headers (max-age=31536000, immutable)"
            fi
            ;;
        "medium-term")
            if [[ "$cache_control" == *"max-age=2592000"* ]]; then
                print_success "Medium-term caching headers are correct"
            else
                print_error "Expected medium-term caching headers (max-age=2592000)"
            fi
            ;;
        "short-term")
            if [[ "$cache_control" == *"max-age=604800"* ]]; then
                print_success "Short-term caching headers are correct"
            else
                print_error "Expected short-term caching headers (max-age=604800)"
            fi
            ;;
    esac

    # Test conditional request with ETag if present
    if [[ -n "$etag" ]]; then
        echo -e "\n${YELLOW}Conditional request with ETag (should receive 304 Not Modified):${NC}"
        conditional_response=$(curl -s -I -H "If-None-Match: ${etag}" "${url}" 2>/dev/null || echo "ERROR")

        if [[ "$conditional_response" != "ERROR" ]]; then
            conditional_status=$(echo "$conditional_response" | head -n1 | cut -d' ' -f2)
            if [[ "$conditional_status" == "304" ]]; then
                print_success "Conditional request returned 304 Not Modified"
            else
                print_warning "Conditional request returned ${conditional_status} (expected 304)"
            fi
        else
            print_error "Failed conditional request"
        fi
    fi

    echo ""
}

# Function to check if server is running
check_server() {
    print_header "Checking server connectivity"

    if curl -s "${SERVER_URL}${SERVER_PATH_PREFIX}/" >/dev/null 2>&1; then
        print_success "Server is running at ${SERVER_URL}${SERVER_PATH_PREFIX}"
        return 0
    else
        print_error "Server is not running at ${SERVER_URL}${SERVER_PATH_PREFIX}"
        echo "Please start the server with: make dev"
        return 1
    fi
}

# Function to show asset versioning
show_asset_versioning() {
    print_header "Asset Versioning Test"

    local css_url="${SERVER_URL}${SERVER_PATH_PREFIX}/css/ui.min.css"

    echo "Making request to CSS file to check versioning:"
    echo "URL: ${css_url}"

    # Check if the asset URL includes version parameter
    response=$(curl -s "${css_url}" 2>/dev/null || echo "ERROR")

    if [[ "$response" != "ERROR" ]]; then
        print_success "CSS file loaded successfully"

        # Show first few lines of CSS to verify it's working
        echo -e "\n${YELLOW}First few lines of CSS:${NC}"
        echo "$response" | head -3

        echo -e "\n${YELLOW}Note:${NC} In the browser, this asset would be requested with a version parameter like:"
        echo "  ${css_url}?v=1705405800"
        echo "  This ensures cache invalidation when the server restarts."
    else
        print_error "Failed to load CSS file"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                  Asset Caching Test Script                  ║"
    echo "║                                                              ║"
    echo "║ This script tests the HTTP cache headers for various assets ║"
    echo "║ to verify that the caching middleware is working correctly. ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Check if server is running
    if ! check_server; then
        exit 1
    fi

    # Test different types of assets
    print_header "Testing Cache Headers for Different Asset Types"

    # CSS files (long-term caching)
    test_cache_headers "/css/ui.min.css" "long-term"
    test_cache_headers "/css/bootstrap-icons.min.css" "long-term"

    # JavaScript files (long-term caching)
    test_cache_headers "/js/htmx-v2.0.6.min.js" "long-term"
    test_cache_headers "/js/htmx-ext-ws-v2.0.3.min.js" "long-term"

    # Images (medium-term caching)
    test_cache_headers "/icon.png" "medium-term"
    test_cache_headers "/apple-touch-icon-180x180.png" "medium-term"

    # Icons (short-term caching)
    test_cache_headers "/favicon.ico" "short-term"

    # Show asset versioning demonstration
    show_asset_versioning

    print_header "Summary"
    print_success "Cache testing completed!"
    echo ""
    echo -e "${YELLOW}What this means for your application:${NC}"
    echo "• CSS and JS files will be cached for 1 year (with version invalidation)"
    echo "• Images will be cached for 30 days"
    echo "• Icons will be cached for 1 week"
    echo "• ETags enable efficient cache validation"
    echo "• Asset versioning uses startup time for automatic cache invalidation"
    echo ""
    echo -e "${BLUE}To see this in action:${NC}"
    echo "1. Open browser developer tools (F12)"
    echo "2. Go to Network tab"
    echo "3. Visit your application"
    echo "4. Refresh the page - assets should show 'from cache'"
    echo "5. Restart server to see new version timestamps in action"
    echo ""
}

# Handle command line arguments
if [[ $# -gt 0 ]]; then
    case "$1" in
        "--help"|"-h")
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -h, --help     Show this help message"
            echo "  --url URL      Override server URL (default: http://localhost:9020)"
            echo "  --prefix PATH  Override server path prefix (default: /pg-press)"
            echo ""
            echo "Make sure your server is running before running this script:"
            echo "  make dev"
            exit 0
            ;;
        "--url")
            SERVER_URL="$2"
            shift 2
            ;;
        "--prefix")
            SERVER_PATH_PREFIX="$2"
            shift 2
            ;;
    esac
fi

# Run main function
main "$@"
