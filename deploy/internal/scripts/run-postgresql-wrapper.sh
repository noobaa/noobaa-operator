#!/bin/bash

# Create diagnostics directory if it doesn't exist
DIAG_DIR="/var/lib/pgsql/diagnostics"
mkdir -p "$DIAG_DIR"

# Create a timestamped log file for this run
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="${DIAG_DIR}/postgres_${TIMESTAMP}.log"

# Function to log messages with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting PostgreSQL wrapper script"
log "Log file: $LOG_FILE"

# Log environment variables (excluding sensitive data)
log "Environment variables:"
env | grep -v -E 'PASSWORD|SECRET|KEY' | sort | tee -a "$LOG_FILE"

# Log system information
log "System information:"
df -h | tee -a "$LOG_FILE"
free -h | tee -a "$LOG_FILE"

log "Starting PostgreSQL initialization..."

# Set up logging for all subsequent commands and execute run-postgresql
exec 1> >(tee -a "$LOG_FILE") 2>&1 /usr/bin/run-postgresql "$@" 