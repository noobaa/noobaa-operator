#!/bin/bash

# This script will automate the file alteration, which involves
# 1. Updating the version of the CLI in version/version.go
# 2. Updating the version of noobaa core image in pkg/options.go => ContainerImageTag
# 3. Run make gen
# 4. Commit the changes with an automated commit message

set -eux
set -o pipefail

export PS4='\e[36m+ ${FUNCNAME:-main}@${BASH_SOURCE}:${LINENO} \e[0m'

dir=$(dirname "$0")

# Source the utilities
source $dir/utils.sh

version=$(bump_semver_patch $(get_noobaa_version))
DRY_RUN=${DRY_RUN:="false"}

# Update version of the CLI in version/version.go
function update_version() {
  local version_file=version/version.go
  # Replace the version line with the new version using perl because sed is not compatible with both mac and linux
  # and awk solution is clumsy
  finline_replace "Version = \".*\"" "Version = \"${version}\"" $version_file
}

# Update version of noobaa core image in pkg/options.go => ContainerImageTag
function update_core_container_image_tag() {
  local options_file=pkg/options/options.go
  # Replace version line with new version using perl because sed is not compatible with both mac and linux
  # and awk solution is clumsy
  finline_replace ".*ContainerImageTag = \".*\"" "	ContainerImageTag = \"${version}\"" $options_file
}

# Run make gen
function run_make_gen() {
  make gen
}

# Commit the changes with automated commit message
function commit_changes() {
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY_RUN is set to true: skipping commiting post release changes"
    return
  fi

  git add .
  git commit -m "Automated commit: Bump version to ${version}"
  git push
}

# Main function
function main() {
  update_version
  update_core_container_image_tag
  run_make_gen
  commit_changes
}

main "$@"
