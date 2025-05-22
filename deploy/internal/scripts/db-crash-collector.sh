#!/bin/bash

set -euo pipefail

# Create diagnostics directory if it doesn't exist
DIAG_DIR="/var/lib/pgsql/diagnostics/crash_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$DIAG_DIR"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$DIAG_DIR/collector.log"
}

log "Starting crash diagnostics collection"

# Find the correct userdata directory
USERDATA_DIR=""
if [ -d "/var/lib/pgsql/data/userdata" ]; then
    USERDATA_DIR="/var/lib/pgsql/data/userdata"
elif [ -d "/var/lib/pgsql/data/data/userdata/data/userdata" ]; then
    USERDATA_DIR="/var/lib/pgsql/data/data/userdata/data/userdata"
else
    log "WARNING: Could not find userdata directory in either expected location."
fi

# Save relevant config files if possible
if [ -n "$USERDATA_DIR" ]; then
    log "Using userdata directory: $USERDATA_DIR"
    for f in postgresql.conf pg_hba.conf; do
        if [ -f "$USERDATA_DIR/$f" ]; then
            cp "$USERDATA_DIR/$f" "$DIAG_DIR/"
        fi
    done
    # Copy PostgreSQL logs if present
    if [ -d "$USERDATA_DIR/log" ]; then
        cp -r "$USERDATA_DIR/log" "$DIAG_DIR/pg_log"
    fi
fi

# Collect PostgreSQL database information
log "Collecting PostgreSQL database information"
psql -c "\du" > "$DIAG_DIR/roles.txt" 2>/dev/null || log "Failed to get roles"
psql -c "\l" > "$DIAG_DIR/databases.txt" 2>/dev/null || log "Failed to get databases"
psql -d nbcore -c "\dt" > "$DIAG_DIR/nbcore_tables.txt" 2>/dev/null || log "Failed to get nbcore tables"

# Collect environment variables
log "Collecting environment variables"
env | sort > "$DIAG_DIR/env.txt"

# Collect process list
log "Collecting process list"
ps auxww > "$DIAG_DIR/ps.txt"

# Save termination log if it exists
if [ -f /dev/termination-log ]; then
    log "Saving termination log"
    cp /dev/termination-log "$DIAG_DIR/termination-log.txt"
fi

# Create a summary file
SUMMARY_FILE="$DIAG_DIR/summary.txt"
log "Creating summary file"
echo "Crash Diagnostics Summary" > "$SUMMARY_FILE"
echo "========================" >> "$SUMMARY_FILE"
echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')" >> "$SUMMARY_FILE"
echo "Pod Name: ${HOSTNAME:-unknown}" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "Files collected:" >> "$SUMMARY_FILE"
ls -l "$DIAG_DIR" >> "$SUMMARY_FILE"

log "Crash diagnostics collection complete. Files saved in $DIAG_DIR" 