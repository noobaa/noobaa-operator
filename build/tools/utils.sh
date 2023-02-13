#!/bin/bash

export PS4='\e[36m+ ${FUNCNAME:-main}@${BASH_SOURCE}:${LINENO} \e[0m'

function check_deps() {
  local dependencies=("$@")
  echo "Checking dependencies..."

  for dep in ${dependencies[@]}; do
    if ! command -v $dep &>/dev/null; then
      echo "Please install $dep"
      exit 1
    fi
  done

  echo "All dependencies are installed"
}

function strip_prefix() {
  local str="$1"
  local prefix="$2"

  echo ${str#"$prefix"}
}

function version_without_v() {
  strip_prefix "$1" v
}

function version_with_v() {
  local without_v=$(version_without_v "$1")

  echo "v$without_v"
}

# get_noobaa_version returns the NooBaa version by reading the `cmd/version/main.go`
#
# The function can be called without any arguments in which case it will return
# noobaa version without "v" prefixed to the version however if ANY second
# argument is provided to the function then it will prefix "v" to the version.
#
# Example: get_noobaa_version # returns version without "v" prefix
# Example: get_noobaa_version 1 # returns version with "v" prefix
function get_noobaa_version() {
  local version=$(go run cmd/version/main.go)
  if [[ $# -gte 1 ]]; then
    version_with_v "$version"
  else
    version_without_v "$version"
  fi
}

# finline_replace takes 3 arguments. First is the regex for the line which
# needs to be replaced, second is the regex for the line which needs to be
# replace the previous line and third is the name of the file
#
# NOTE: Function relies on perl
function finline_replace() {
  perl -i -pe "s/$1/$2/" "$3"
}

# bump_semver_patch takes in a semver and bumps the patch version
function bump_semver_patch() {
  local version=$(version_without_v "$1")
  version=$(echo ${version} | awk -F. -v OFS=. '{$NF = $NF+1 ; print}')
  echo "$version"
}
