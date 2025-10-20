#!/bin/bash

# Script to monitor and log user activity on specific pages from pg-press service logs
# Tracks activity on: /, /feed, /profile, /editor, /help, /trouble-reports, /notes, /tools

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Define the pages we want to track
TRACKED_PAGES=(
    "/"
    "/feed"
    "/profile"
    "/editor"
    "/help"
    "/trouble-reports"
    "/notes"
    "/tools"
)

# Function to check if a path should be tracked
is_tracked_page() {
    local path="$1"
    # Remove /pg-press prefix if present
    local clean_path="${path#/pg-press}"

    # Handle root path special case
    if [[ "$clean_path" == "/" ]] || [[ "$path" == "/pg-press/" ]]; then
        return 0
    fi

    # Check other paths
    for page in "${TRACKED_PAGES[@]}"; do
        if [[ "$page" != "/" ]] && [[ "$clean_path" == "$page"* ]]; then
            return 0
        fi
    done
    return 1
}

# Function to parse and format log line
parse_log_line() {
    local line="$1"

    # Skip non-request lines
    if [[ ! "$line" =~ \[Server\] ]]; then
        return
    fi

    # Extract components using regex
    # Pattern: ✅ DATE TIME [Server] STATUS METHOD PATH (IP) DURATION User{ID: NUM, Name: NAME}
    if [[ "$line" =~ ([0-9]{4}/[0-9]{2}/[0-9]{2}\ [0-9]{2}:[0-9]{2}:[0-9]{2}).*\[Server\]\ ([0-9]+)\ ([A-Z]+)[[:space:]]+([^[:space:]]+)\ \(([^)]+)\)\ ([^[:space:]]+)\ User\{ID:\ ([0-9]+),\ Name:\ ([^}]+)\} ]]; then
        local datetime="${BASH_REMATCH[1]}"
        local status="${BASH_REMATCH[2]}"
        local method="${BASH_REMATCH[3]}"
        local path="${BASH_REMATCH[4]}"
        local ip="${BASH_REMATCH[5]}"
        local duration="${BASH_REMATCH[6]}"
        local user_id="${BASH_REMATCH[7]}"
        local user_name="${BASH_REMATCH[8]}"

        # Only process successful requests (200 status)
        if [[ "$status" == "200" ]] && is_tracked_page "$path"; then
            # Clean up the path for display
            local display_path="${path#/pg-press}"
            if [[ -z "$display_path" ]]; then
                display_path="/"
            fi

            # Format output
            echo -e "${GREEN}[${datetime}]${NC} ${CYAN}User ID: ${user_id}${NC} | ${YELLOW}${user_name}${NC} | ${BLUE}${display_path}${NC}"
        fi
    fi
}

# Function to display help
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Monitor user activity on specific pages from pg-press service logs"
    echo ""
    echo "Options:"
    echo "  -f, --follow      Follow log output (like tail -f)"
    echo "  -n NUM            Show last NUM lines of activity (default: all)"
    echo "  -s, --since TIME  Show entries since TIME (e.g., '1 hour ago', 'today')"
    echo "  -u, --until TIME  Show entries until TIME"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Tracked pages:"
    for page in "${TRACKED_PAGES[@]}"; do
        echo "  - $page"
    done
    echo ""
    echo "Examples:"
    echo "  $0 -f                    # Follow live activity"
    echo "  $0 -n 50                 # Show last 50 activities"
    echo "  $0 -s '1 hour ago'       # Show activity from last hour"
    echo "  $0 -s today -f           # Follow today's activity"
}

# Parse command line arguments
FOLLOW=false
LINES=""
SINCE=""
UNTIL=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--follow)
            FOLLOW=true
            shift
            ;;
        -n)
            LINES="$2"
            shift 2
            ;;
        -s|--since)
            SINCE="$2"
            shift 2
            ;;
        -u|--until)
            UNTIL="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Build journalctl command
JOURNAL_CMD="journalctl --user -u pg-press --output cat"

if [[ -n "$SINCE" ]]; then
    JOURNAL_CMD="$JOURNAL_CMD --since \"$SINCE\""
fi

if [[ -n "$UNTIL" ]]; then
    JOURNAL_CMD="$JOURNAL_CMD --until \"$UNTIL\""
fi

if [[ "$FOLLOW" == true ]]; then
    JOURNAL_CMD="$JOURNAL_CMD -f"
elif [[ -n "$LINES" ]]; then
    JOURNAL_CMD="$JOURNAL_CMD -n $LINES"
else
    JOURNAL_CMD="$JOURNAL_CMD --no-pager"
fi

# Header
echo -e "${GREEN}=== PG-Press User Activity Monitor ===${NC}"
echo -e "Tracking pages: ${TRACKED_PAGES[*]}"
echo -e "${GREEN}=====================================${NC}"
echo ""

# Process logs
if [[ "$FOLLOW" == true ]]; then
    # Follow mode - process lines as they come
    eval "$JOURNAL_CMD" | while IFS= read -r line; do
        parse_log_line "$line"
    done
else
    # Batch mode - collect and process all lines
    activity_found=false
    while IFS= read -r line; do
        result=$(parse_log_line "$line")
        if [[ -n "$result" ]]; then
            echo "$result"
            activity_found=true
        fi
    done < <(eval "$JOURNAL_CMD")

    if [[ "$activity_found" == false ]]; then
        echo "No user activity found for the specified criteria."
    fi
fi
