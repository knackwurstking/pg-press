#!/bin/bash

# Script to monitor and log user activity on specific pages from pg-press service logs
# Tracks activity on: /, /feed, /profile, /editor, /help, /trouble-reports, /notes, /tools
# Version: 2.0 - Fixed regex pattern for log parsing

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Define the pages we want to track (without /pg-press prefix)
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
    # Remove /pg-press prefix if present (handle both with and without trailing slash)
    local clean_path="${path#/pg-press}"

    # If nothing was removed or only "/" remains, it's the root
    if [[ -z "$clean_path" ]] || [[ "$clean_path" == "/" ]]; then
        clean_path="/"
    fi

    if [[ "$DEBUG" == true ]]; then
        echo -e "${YELLOW}[DEBUG] is_tracked_page: original path='$path', clean_path='$clean_path'${NC}" >&2
    fi

    # Check if the clean path matches any tracked page
    for page in "${TRACKED_PAGES[@]}"; do
        if [[ "$DEBUG" == true ]]; then
            echo -e "${YELLOW}[DEBUG] is_tracked_page: checking against tracked page='$page'${NC}" >&2
        fi
        if [[ "$page" == "/" ]]; then
            # Root path special case
            if [[ "$clean_path" == "/" ]]; then
                if [[ "$DEBUG" == true ]]; then
                    echo -e "${GREEN}[DEBUG] is_tracked_page: MATCHED root path${NC}" >&2
                fi
                return 0
            fi
        else
            # For non-root paths, check if clean_path starts with the tracked page
            if [[ "$clean_path" == "$page" ]] || [[ "$clean_path" == "$page/"* ]]; then
                if [[ "$DEBUG" == true ]]; then
                    echo -e "${GREEN}[DEBUG] is_tracked_page: MATCHED '$page'${NC}" >&2
                fi
                return 0
            fi
        fi
    done
    if [[ "$DEBUG" == true ]]; then
        echo -e "${RED}[DEBUG] is_tracked_page: NO MATCH found${NC}" >&2
    fi
    return 1
}

# Function to parse and format log line
parse_log_line() {
    local line="$1"

    # Debug mode: show raw line
    if [[ "$DEBUG" == true ]]; then
        echo -e "${YELLOW}[DEBUG] Raw line: ${NC}$line"
    fi

    # Skip non-request lines
    if [[ ! "$line" =~ \[Server\] ]]; then
        if [[ "$DEBUG" == true ]]; then
            echo -e "${YELLOW}[DEBUG] Skipped: No [Server] tag found${NC}"
        fi
        return
    fi

    # Extract components using regex
    # Pattern: ✅ DATE TIME [Server] STATUS METHOD PATH (IP) DURATION User{ID: NUM, Name: NAME}
    # Using a simpler, more explicit regex pattern
    local full_regex='([0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}) \[Server\] ([0-9]+) ([A-Z]+)[[:space:]]+([^[:space:]]+) \(([^)]+)\) ([^[:space:]]+) User\{ID: ([0-9]+), Name: ([^}]+)\}'

    if [[ "$DEBUG" == true ]]; then
        echo -e "${YELLOW}[DEBUG] Regex: ${NC}$full_regex"
        echo -e "${YELLOW}[DEBUG] Testing match...${NC}"
    fi

    if [[ "$line" =~ $full_regex ]]; then
        local datetime="${BASH_REMATCH[1]}"
        local status="${BASH_REMATCH[2]}"
        local method="${BASH_REMATCH[3]}"
        local path="${BASH_REMATCH[4]}"
        local ip="${BASH_REMATCH[5]}"
        local duration="${BASH_REMATCH[6]}"
        local user_id="${BASH_REMATCH[7]}"
        local user_name="${BASH_REMATCH[8]}"

        if [[ "$DEBUG" == true ]]; then
            echo -e "${GREEN}[DEBUG] Match found!${NC}"
            echo -e "${YELLOW}[DEBUG] Extracted: datetime=$datetime, status=$status, method=$method, path=$path, user_id=$user_id, user_name=$user_name${NC}"
        fi

        # Only process successful requests (200 status)
        if [[ "$status" == "200" ]] && is_tracked_page "$path"; then
            # Clean up the path for display
            local display_path="${path#/pg-press}"
            if [[ -z "$display_path" ]]; then
                display_path="/"
            fi

            # Format output
            echo -e "${GREEN}[${datetime}]${NC} ${CYAN}User ID: ${user_id}${NC} | ${YELLOW}${user_name}${NC} | ${BLUE}${display_path}${NC}"
        elif [[ "$DEBUG" == true ]]; then
            echo -e "${YELLOW}[DEBUG] Filtered out: status=$status, tracked=$(is_tracked_page "$path" && echo "yes" || echo "no")${NC}"
        fi
    elif [[ "$DEBUG" == true ]]; then
        echo -e "${RED}[DEBUG] No regex match!${NC}"
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
    echo "  -l, --log FILE    Read from log file instead of journalctl"
    echo "  -d, --debug       Enable debug mode to show raw lines and parsing info"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Tracked pages:"
    for page in "${TRACKED_PAGES[@]}"; do
        echo "  - $page"
    done
    echo ""
    echo "Examples:"
    echo "  $0 -f                    # Follow live activity (journalctl)"
    echo "  $0 -n 50                 # Show last 50 activities"
    echo "  $0 -s '1 hour ago'       # Show activity from last hour"
    echo "  $0 -s today -f           # Follow today's activity"
    echo "  $0 -l server.log         # Read from log file"
    echo "  $0 -l server.log -f      # Follow log file"
    echo "  $0 -d                    # Debug mode to troubleshoot parsing"
    echo ""
    echo "Note: If journalctl is not available, use the -l option to specify a log file."
}

# Function to check if journalctl is available
check_journalctl() {
    if command -v journalctl &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to detect the correct service name and scope
detect_service() {
    local service_variations=("pg-press" "pgpress" "pg_press")
    local scopes=("--user" "")

    for scope in "${scopes[@]}"; do
        for service in "${service_variations[@]}"; do
            # Try to get one line from the service
            if journalctl $scope -u "$service" -n 1 &>/dev/null; then
                echo "$scope|$service"
                return 0
            fi
        done
    done

    # Nothing found
    return 1
}

# Parse command line arguments
FOLLOW=false
LINES=""
SINCE=""
UNTIL=""
LOG_FILE=""
DEBUG=false

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
        -l|--log)
            LOG_FILE="$2"
            shift 2
            ;;
        -d|--debug)
            DEBUG=true
            shift
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

# Header
echo -e "${GREEN}=== PG-Press User Activity Monitor ===${NC}"
echo -e "Tracking pages: ${TRACKED_PAGES[*]}"
echo -e "${GREEN}=====================================${NC}"
echo ""

# Determine input source and process logs
if [[ -n "$LOG_FILE" ]]; then
    # Read from log file
    if [[ ! -f "$LOG_FILE" ]]; then
        echo -e "${RED}Error: Log file '$LOG_FILE' does not exist.${NC}"
        exit 1
    fi

    echo -e "Reading from log file: ${CYAN}${LOG_FILE}${NC}"
    echo ""

    if [[ "$FOLLOW" == true ]]; then
        # Follow mode for log file
        tail -f "$LOG_FILE" | while IFS= read -r line; do
            parse_log_line "$line"
        done
    else
        # Read entire file or last N lines
        if [[ -n "$LINES" ]]; then
            tail -n "$LINES" "$LOG_FILE" | while IFS= read -r line; do
                parse_log_line "$line"
            done
        else
            activity_found=false
            while IFS= read -r line; do
                result=$(parse_log_line "$line")
                if [[ -n "$result" ]]; then
                    echo "$result"
                    activity_found=true
                fi
            done < "$LOG_FILE"

            if [[ "$activity_found" == false ]]; then
                echo "No user activity found in the log file."
            fi
        fi
    fi
else
    # Try to use journalctl
    if ! check_journalctl; then
        echo -e "${RED}Error: journalctl is not available on this system.${NC}"
        echo "Please use the -l option to specify a log file."
        echo ""
        echo "Example: $0 -l /path/to/pg-press.log"
        exit 1
    fi

    # Detect the correct service name and scope
    echo -e "${CYAN}Detecting pg-press service...${NC}"
    SERVICE_INFO=$(detect_service)

    if [[ $? -ne 0 ]]; then
        echo -e "${RED}Error: Could not find pg-press service in journalctl.${NC}"
        echo "Tried variations: pg-press, pgpress, pg_press (in both user and system scope)"
        echo ""
        echo "Please check if the service is running or use the -l option with a log file."
        echo "Example: $0 -l /path/to/pg-press.log"
        exit 1
    fi

    # Parse service info
    IFS='|' read -r SCOPE SERVICE_NAME <<< "$SERVICE_INFO"

    # Build journalctl command
    JOURNAL_CMD="journalctl $SCOPE -u $SERVICE_NAME --output cat"

    if [[ "$DEBUG" == true ]]; then
        echo -e "${GREEN}[DEBUG] Detected service: $SERVICE_NAME (scope: ${SCOPE:-system})${NC}"
    fi

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

    echo -e "Reading from: ${CYAN}journalctl ($SERVICE_NAME${SCOPE:+ }${SCOPE#--})${NC}"

    if [[ "$DEBUG" == true ]]; then
        echo -e "${YELLOW}[DEBUG] Running: $JOURNAL_CMD${NC}"
    fi
    echo ""

    # Process logs
    if [[ "$FOLLOW" == true ]]; then
        # Follow mode - process lines as they come
        eval "$JOURNAL_CMD" 2>/dev/null | while IFS= read -r line; do
            parse_log_line "$line"
        done

        # Check if journalctl command failed
        if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
            echo -e "${RED}Error: Failed to read from journalctl. The service might not exist or you might not have permission.${NC}"
            echo "Try using sudo or check if the pg-press service is installed."
            exit 1
        fi
    else
        # Batch mode - collect and process all lines
        activity_found=false

        # Capture journalctl output and check for errors
        journal_output=$(eval "$JOURNAL_CMD" 2>&1)
        journal_status=$?

        if [[ $journal_status -ne 0 ]]; then
            echo -e "${RED}Error: Failed to read from journalctl.${NC}"
            echo "Output: $journal_output"
            echo ""
            echo "The service might not exist or you might not have permission."
            echo "Try using the -l option with a log file instead."
            exit 1
        fi

        while IFS= read -r line; do
            result=$(parse_log_line "$line")
            if [[ -n "$result" ]]; then
                echo "$result"
                activity_found=true
            fi
        done <<< "$journal_output"

        if [[ "$activity_found" == false ]]; then
            echo "No user activity found for the specified criteria."
        fi
    fi
fi
