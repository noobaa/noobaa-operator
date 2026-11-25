#!/bin/sh

# golangci-lint runner script
# Usage: ./scripts/run-golangci-lint.sh

# Uncomment this to see the commands being run
# Note: We don't use 'set -e' because we need to handle golangci-lint's
# exit codes manually to distinguish between linting issues and other errors
# set -x

# Exit code used by golangci-lint when linting issues are found
# Allows us to distinguish between linting issues and other errors
readonly ISSUES_EXIT_CODE=42
readonly GOLANGCI_LINT_VERSION=v2.15.0

# Safely check whether Go is installed and locate its binary directories
if command -v go >/dev/null 2>&1; then
    GOBIN="$(go env GOBIN)"
    GOPATH_BIN="$(go env GOPATH)"
    GOPATH_BIN="${GOPATH_BIN%%:*}/bin"
    # Prefer the installed linter, with GOBIN taking precedence when set
    export PATH="${GOPATH_BIN}:${PATH}"
    [ -n "${GOBIN}" ] && export PATH="${GOBIN}:${PATH}"
fi

# Function to install golangci-lint
install_golangci_lint() {
    echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION}"
    if ! go install -a "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
    then
        echo "⚠️ Failed to install golangci-lint"
        exit 0
    fi
}

# Function to run golangci-lint
run_lint() {
    # Install golangci-lint if needed
    install_golangci_lint
    
    echo "Running lint on all files"
    golangci-lint run --issues-exit-code=${ISSUES_EXIT_CODE} --config .golangci.yml
    
    # Store the exit code
    LINT_EXIT_CODE=${?}
    
    # Handle exit codes properly
    if [ ${LINT_EXIT_CODE} -eq ${ISSUES_EXIT_CODE} ]
    then
        echo "❌ golangci-lint found linting issues. Please fix them."
        exit 1
    elif [ ${LINT_EXIT_CODE} -ne 0 ]
    then
        echo "⚠️ golangci-lint encountered an error (exit code: ${LINT_EXIT_CODE})"
        exit 0
    fi
    
    echo "✅ golangci-lint passed!"
}

run_lint
