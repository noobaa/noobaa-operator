#!/bin/sh

# Unified golangci-lint runner script
# Usage: ./scripts/run-golangci-lint.sh <mode>
# Modes: precommit, makefile

# Uncomment this to see the commands being run
# Note: We don't use 'set -e' because we need to handle golangci-lint's
# exit codes manually to distinguish between linting issues and other errors
# set -x

# Exit code used by golangci-lint when linting issues are found
# Allows us to distinguish between linting issues and other errors
readonly ISSUES_EXIT_CODE=42

MODE="${1}"

# Safely check whether Go is installed and set the GOPATH
# Also check for GOBIN and add it to the PATH in case the binary was installed to a non-standard location
 if command -v go >/dev/null 2>&1; then
     GOBIN="$(go env GOBIN)"
     GOPATH_BIN="$(go env GOPATH)/bin"
     [ -n "${GOBIN}" ] && export PATH="${PATH}:${GOBIN}"
     export PATH="${PATH}:${GOPATH_BIN}"
 fi

# Function to install golangci-lint (for makefile mode)
install_golangci_lint() {
    echo "Installing the latest golangci-lint with local toolchain"
    if ! go install -a "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0"
    then
        echo "⚠️ Failed to install golangci-lint"
        exit 0
    fi
}

# Function to run golangci-lint in precommit mode
run_precommit_lint() {
    echo "Running golangci-lint in precommit mode..."
    
    # Check if golangci-lint is available - exit gracefully if it isn't
    if ! command -v golangci-lint >/dev/null 2>&1
    then
        echo "⚠️ golangci-lint not found – run 'make lint' to install it automatically"
        exit 0
    fi
    
    # Collect staged Go files as positional args
    # cached: retrieve only staged files
    # name-only: retrieve only the names of the files
    # diff-filter=ACMR: retrieve only the files that have been added, copied, modified, or renamed
    # grep -E '\.go$': retrieve only the files that end with .go
    # || true: if there are no staged Go files, return a zero-status exit code
    set -- $(git diff --cached --name-only --diff-filter=ACMR | grep -E '\.go$' || true)
    # If there are no staged Go files, exit successfully
    if [ "$#" -eq 0 ]; then
        echo "✅ No staged Go files were found, exiting"
        exit 0
    fi
    
    echo "Running golangci-lint on staged files..."
    BASE=$(git rev-parse --verify HEAD 2>/dev/null || echo "")
    
    if [ -z "${BASE}" ]
    then
        # Initial commit – lint only staged files
        golangci-lint run --issues-exit-code=${ISSUES_EXIT_CODE} --config .golangci.yml "${@}"
    else
        # Lint only staged changes vs HEAD
        golangci-lint run --issues-exit-code=${ISSUES_EXIT_CODE} --config .golangci.yml --new-from-rev="${BASE}" "${@}"
    fi
    
    # Store the exit code
    LINT_EXIT_CODE=${?}
    
    # Handle exit codes properly
    if [ ${LINT_EXIT_CODE} -eq ${ISSUES_EXIT_CODE} ]
    then
        echo "❌ golangci-lint found linting issues in staged files. Please fix them before committing."
        exit 1
    elif [ ${LINT_EXIT_CODE} -ne 0 ]
    then
        echo "⚠️  golangci-lint encountered an error (exit code: ${LINT_EXIT_CODE})"
        exit 0
    fi
    
    echo "✅ golangci-lint passed!"
}

# Function to run golangci-lint in makefile mode
run_makefile_lint() {
    echo "Running golangci-lint in makefile mode..."
    
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

# Main logic
case "${MODE}" in
    "precommit")
        run_precommit_lint
        ;;
    "makefile")
        run_makefile_lint
        ;;
    *)
        echo "Usage: ${0} <mode>"
        echo "Modes: precommit, makefile"
        exit 1
        ;;
esac
