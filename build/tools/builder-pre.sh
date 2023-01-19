#!/bin/bash

# NOTE: This script can be run manually but is run anyway by builder.sh before that script starts executing its `main`

# This script will automate the file alteration, which involves
# 1. Updating the README.md
# 2. Committing the changes and pushing the changes

set -eux
set -o pipefail

export PS4='\e[36m+ ${FUNCNAME:-main}@${BASH_SOURCE}:${LINENO} \e[0m'

dir=$(dirname "$0")

# Source the utilities
source $dir/utils.sh

# Variables
DRY_RUN=${DRY_RUN:="false"}

# Update the version of noobaa core image, noobaa operator image and CLI version in the README output.
function update_readme() {
	local version=$(get_noobaa_version)
	local core_version="$version" # Assume to be the same as the version of operator
	local readme_file=README.md

	finline_replace "INFO\[0000\] CLI version: .*" "INFO\[0000\] CLI version: ${version}" $readme_file
	finline_replace "INFO\[0000\] noobaa-image: .*" "INFO\[0000\] noobaa-image: noobaa\/noobaa-core:${core_version}" $readme_file
	finline_replace "INFO\[0000\] operator-image: .*" "INFO\[0000\] operator-image: noobaa\/noobaa-operator:${version}" $readme_file
}

function commit_changes() {
	if [[ $DRY_RUN == "true" ]]; then
		echo "DRY_RUN is set to true, skipping commiting changes."
		return
	fi

	local version=$(get_noobaa_version 1)

	git add .
	git commit -m "Automated commit to update README for version: ${version}"
	git push

	# If TAG=1 is provided, then create a tag
	if [ -n "$TAG" ]; then
		git tag -a "${version}" -m "Tag for version ${version}"
		git push origin "${version}"
	fi
}

# Main function
function main() {
	update_readme
	commit_changes
}

main "$@"
