#!/bin/bash

set -euox pipefail

dir=$(dirname "$0")

export PS4='\e[36m+ ${FUNCNAME:-main}@${BASH_SOURCE}:${LINENO} \e[0m'

# Source the utilities
source $dir/utils.sh

dependencies=(docker sha256sum git gh yq)

# Default values for the arguments.
GH_ORG="noobaa"
GH_REPO="noobaa-operator"
OCI_ORG="noobaa"
DRY_RUN="false"
DRAFT="false"
ONLY_GH="false"
ONLY_OCI="false"
ONLY_HOMEBREW="false"
ONLY_OPERATOR_HUB="false"

# Default values for the environment variables.
ARTIFACTS_PATH=${ARTIFACTS_PATH:-$dir/artifacts}
GITHUB_TOKEN=${GITHUB_TOKEN:-}
DOCKERHUB_USERNAME=${DOCKERHUB_USERNAME:-}
DOCKERHUB_TOKEN=${DOCKERHUB_TOKEN:-}
QUAY_USERNAME=${QUAY_USERNAME:-}
QUAY_TOKEN=${QUAY_TOKEN:-}
HOMEBREW_CORE_REPO=${HOMEBREW_CORE_REPO:-}
OPERATOR_HUB_REPO=${OPERATOR_HUB_REPO:-}

function check_environment_variables() {
  if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "GITHUB_TOKEN environment variable is not set."
    exit 1
  fi
  if [[ -z "${DOCKERHUB_USERNAME}" ]]; then
    echo "DOCKERHUB_USERNAME environment variable is not set."
    exit 1
  fi
  if [[ -z "${DOCKERHUB_TOKEN}" ]]; then
    echo "DOCKERHUB_TOKEN environment variable is not set."
    exit 1
  fi
  if [[ -z "${QUAY_USERNAME}" ]]; then
    echo "QUAY_USERNAME environment variable is not set."
    exit 1
  fi
  if [[ -z "${QUAY_TOKEN}" ]]; then
    echo "QUAY_TOKEN environment variable is not set."
    exit 1
  fi
  if [[ -z "${HOMEBREW_CORE_REPO}" ]]; then
    echo "HOMEBREW_CORE_REPO environment variable is not set."
    exit 1
  fi
  if [[ -z "${ARTIFACTS_PATH}" ]]; then
    echo "ARTIFACTS_PATH environment variable is not set."
    exit 1
  fi
  if [[ -z "${OPERATOR_HUB_REPO}" ]]; then
    echo "OPERATOR_HUB_REPO environment variable is not set."
    exit 1
  fi
}

function parse_args() {
  # Parse command-line arguments.
  while [ "$#" -gt 0 ]; do
    case "$1" in
    --gh-org)
      GH_ORG="$2"
      shift
      ;;
    --gh-repo)
      GH_REPO="$2"
      shift
      ;;
    --oci-org)
      OCI_ORG="$2"
      shift
      ;;
    --only-gh)
      ONLY_GH="true"
      ;;
    --only-oci)
      ONLY_OCI="true"
      ;;
    --only-homebrew)
      ONLY_HOMEBREW="true"
      ;;
    --only-operator-hub)
      ONLY_OPERATOR_HUB="true"
      ;;
    --dry-run)
      DRY_RUN="true"
      ;;
    --draft)
      DRAFT="true"
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 1
      ;;
    esac
    shift
  done
}

function prepare_release() {
  local os="$1"
  local arch="$2"
  local version=$(get_noobaa_version 1)

  # Create a temporary directory to store the release files.
  local release_dir=$(mktemp -d)
  echo "Created a temporary directory for the release: ${release_dir}"

  cp "$ARTIFACTS_PATH/bin/noobaa-$version-$os-$arch" "$release_dir/noobaa-operator"
  cp "LICENSE" "$release_dir/LICENSE"

  # Create the release archive.
  local release_archive="$ARTIFACTS_PATH/release/noobaa-operator-$version-$os-$arch.tar.gz"
  tar -czf "$release_archive" -C "$release_dir" .
  echo "Created the release archive: ${release_archive}"

  # Create the release checksum.
  local release_checksum="noobaa-operator-$version-$os-$arch.tar.gz.sha256"
  (
    cd $ARTIFACTS_PATH/release
    sha256sum "noobaa-operator-$version-$os-$arch.tar.gz" >"$release_checksum"
  )
  echo "Created the release checksum: ${release_checksum}"

  # Remove the temporary directory.
  rm -rf "$release_dir"
}

function create_gh_release() {
  # Generate all the release files.
  prepare_release "linux" "amd64"
  prepare_release "darwin" "amd64"
  prepare_release "linux" "arm64"
  prepare_release "darwin" "arm64"

  # Create the release on GitHub using gh and github REST API
  # https://cli.github.com/manual/gh_release_create
  # https://docs.github.com/en/rest/reference/repos#create-a-release
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY_RUN is set, skipping Github release creation."
    return
  fi

  # Create the release on GitHub.
  echo "Creating the release on GitHub..."
  gh api \
    --method POST \
    -H "Accept: application/vnd.github+json" \
    /repos/$GH_ORG/$GH_REPO/releases \
    -f tag_name="$(get_noobaa_version 1)" \
    -f name="$(get_noobaa_version 1)" \
    -f body="Release $(get_noobaa_version 1)" \
    -F draft="$DRAFT" \
    -F generate_release_notes="true"

  # Upload the release files to GitHub.
  echo "Uploading the release files to GitHub..."
  gh release upload "$(get_noobaa_version 1)" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-linux-amd64.tar.gz" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-linux-amd64.tar.gz.sha256" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-darwin-amd64.tar.gz" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-darwin-amd64.tar.gz.sha256" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-linux-arm64.tar.gz" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-linux-arm64.tar.gz.sha256" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-darwin-arm64.tar.gz" \
    "$ARTIFACTS_PATH/release/noobaa-operator-$(get_noobaa_version 1)-darwin-arm64.tar.gz.sha256"
}

function create_oci_release() {
  # Release OCI images to docker and quay.io
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY_RUN is set, skipping OCI release creation."
    return
  fi

  local quay_image="quay.io/$QUAY_USERNAME/noobaa-operator:$(get_noobaa_version)"
  local docker_image="$DOCKERHUB_USERNAME/noobaa-operator:$(get_noobaa_version)"

  echo "Tagging the images..."
  docker tag "$OCI_ORG/noobaa-operator:$(get_noobaa_version)" $quay_image
  docker tag "$OCI_ORG/noobaa-operator:$(get_noobaa_version)" $docker_image

  echo "Logging in to docker.io..."
  echo "$DOCKERHUB_TOKEN" | docker login -u "$DOCKERHUB_USERNAME" --password-stdin

  echo "Pushing the images to docker.io..."
  docker push $docker_image

  echo "Logging in to quay.io..."
  echo "$QUAY_TOKEN" | docker login -u "$QUAY_USERNAME" --password-stdin quay.io

  echo "Pushing the images to quay.io..."
  docker push $quay_image
}

function create_krew() {
  # Rely on the krew-release-bot to create the krew release
  \cp "$ARTIFACTS_PATH/krew/noobaa.yaml" "./.krew.yaml"

  return
}

function release_homebrew() {
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY_RUN is set, skipping homebrew release creation."
    return
  fi

  local version=$(get_noobaa_version 1)

  local homebrew_dir=$(mktemp -d)
  echo "Created a temporary directory for homebrew-core: ${homebrew_dir}"

  echo "Cloning homebrew-core..."
  git clone https://$GITHUB_TOKEN@github.com/$HOMEBREW_CORE_REPO $homebrew_dir

  echo "Updating the manifest..."
  mkdir -p "$homebrew_dir/Formula"
  cp "$ARTIFACTS_PATH/homebrew/noobaa.rb" "$homebrew_dir/Formula/noobaa.rb"

  pushd $homebrew_dir

  echo "Committing the changes..."
  git add "Formula/noobaa.rb"
  git commit -s -m "noobaa-operator $version"

  echo "Pushing the changes..."
  git push -u origin master

  popd

  echo "Cleaning up..."
  rm -rf "$homebrew_dir"
}

function release_operatorhub() {
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY_RUN is set, skipping OLM release creation."
    return
  fi

  local version=$(get_noobaa_version)

  local operatorhub_dir=$(mktemp -d)
  echo "Created a temporary directory for operatorhub.io: ${operatorhub_dir}"

  echo "Cloning operatorhub.io..."
  git clone https://$GITHUB_TOKEN@github.com/$OPERATOR_HUB_REPO $operatorhub_dir

  pushd $operatorhub_dir

  echo "Forking operatorhub.io..."
  gh repo fork --remote
  git remote set-url origin $(git remote get-url origin | sed -r "s/https:\/\/(.*)/https:\/\/$GITHUB_TOKEN\@\1/")

  echo "Updating the manifest..."
  git checkout -b "noobaa-operator-$version"

  local last_version=$(cd operators/noobaa-operator && printf "%s\n" */ | sort -V | sed -e 's-/$--' | tail -n 1)
  popd

  mkdir -p "$operatorhub_dir/operators/noobaa-operator/"
  # \cp is to bypass aliasing of cp to cp -i
  \cp -rf $ARTIFACTS_PATH/operator-bundle/* "$operatorhub_dir/operators/noobaa-operator/"

  pushd $operatorhub_dir

  pushd "operators/noobaa-operator/"
  # Update the ".spec.replaces" field in the CSV
  yq -i ".spec.replaces = \"noobaa-operator.v$last_version\"" "$version/noobaa-operator.v$version.clusterserviceversion.yaml"
  popd

  echo "Committing the changes..."
  git add "operators/noobaa-operator/"
  git commit -s -m "noobaa-operator $version"

  echo "Pushing the changes..."
  git push -u origin "noobaa-operator-$version"

  echo "Creating a pull request..."
  gh pr create --title "noobaa-operator $version" --body "This is an automated PR for the release of Noobaa version: $version"

  popd

  echo "Cleaning up..."
  rm -rf "$operatorhub_dir"
}

function usage() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  --help - Print this help message."
  echo "  --oci-org <org> - The organization of the OCI images."
  echo "  --gh-org <org> - The organization of the GitHub repository."
  echo "  --gh-repo <repo> - The GitHub repository name."
  echo "  --draft - Create the release as a draft."
  echo "  --dry-run - Do not create the release."
  echo "  --only-gh - Create the GitHub release only."
  echo "  --only-oci - Create the OCI release only."
  echo "  --only-homebrew - Create the homebrew release only."
  echo "  --only-operator-hub - Create the operatorhub.io release only."
}

function init() {
  check_environment_variables
  mkdir -p "$ARTIFACTS_PATH/release"
  parse_args "$@"
}

function main() {
  check_deps "${dependencies[@]}"

  # If none of --only-* flags are set, run all the release steps.
  if [[ "$ONLY_GH" == "false" && "$ONLY_OCI" == "false" && "$ONLY_HOMEBREW" == "false" && "$ONLY_OPERATOR_HUB" == "false" ]]; then
    create_gh_release
    create_oci_release
    create_krew
    release_homebrew
    release_operatorhub
  fi

  if [[ "$ONLY_GH" == "true" ]]; then
    create_gh_release
  fi

  if [[ "$ONLY_OCI" == "true" ]]; then
    create_oci_release
  fi

  if [[ "$ONLY_HOMEBREW" == "true" ]]; then
    release_homebrew
  fi

  if [[ "$ONLY_OPERATOR_HUB" == "true" ]]; then
    release_operatorhub
  fi
}

init "$@"

if [ -f "$dir/releaser-pre.sh" ]; then
  bash "$dir/releaser-pre.sh"
fi

main "$@"

if [ -f "$dir/releaser-post.sh" ]; then
  bash "$dir/releaser-post.sh"
fi
