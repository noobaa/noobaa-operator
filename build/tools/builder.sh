#!/bin/bash

# This script is supposed to automate various types of artifact generation that this project has
# to do.
# Following types of artifacts are supported:
# 1. Build a new version of the CLI
# 2. Build a new version of OCI image
# 3. Generate Krew manifest
# 4. Generate Homebrew formula

set -eux
set -o pipefail

export PS4='\e[36m+ ${FUNCNAME:-main}@${BASH_SOURCE}:${LINENO} \e[0m'

dir=$(dirname "$0")

# Source the utilities
source $dir/utils.sh

bin_path=$dir/artifacts/bin
krew_path=$dir/artifacts/krew
homebrew_path=$dir/artifacts/homebrew
operator_bundle_path=$dir/artifacts/operator-bundle
dependencies=(go docker make sha256sum git)
GH_ORG="noobaa"
GH_REPO="noobaa-operator"
DRY_RUN="false"

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
    --dry-run)
      DRY_RUN="true"
      ;;
    -h|--help)
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

function init() {
  mkdir -p $bin_path
  mkdir -p $krew_path
  mkdir -p $homebrew_path
  mkdir -p $operator_bundle_path
  parse_args "$@"
}

function generate_full_bin_path() {
  local GOOS=$1
  local GOARCH=$2

  if [ "$GOOS" == "windows" ]; then
    ext=".exe"
  else
    ext=""
  fi

  echo ${bin_path}/noobaa-$(get_noobaa_version 1)-${GOOS}-${GOARCH}${ext}
}

function build_cli() {
  local GOOS=$1
  local GOARCH=$2

  if [ "$GOOS" == "windows" ]; then
    ext=".exe"
  else
    ext=""
  fi

  GOOS=$GOOS GOARCH=$GOARCH go build -o $(generate_full_bin_path $GOOS $GOARCH) main.go
}

function build_cli_for_all() {
  build_cli linux amd64
  build_cli linux arm64
  build_cli darwin amd64
  build_cli darwin arm64
}

function build_oci() {
  local tag_prefix=$1
  local os=linux
  local arch=amd64

  docker build --platform $os/$arch --build-arg NOOBAA_BIN_PATH=$(generate_full_bin_path $os $arch) -t $tag_prefix/noobaa-operator:$(get_noobaa_version) -f build/Dockerfile .
}

function generate_krew_manifest() {
  cat <<EOF > $krew_path/noobaa.yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: noobaa
spec:
  version: $(get_noobaa_version 1)
  homepage: https://github.com/$GH_ORG/$GH_REPO
  shortDescription: Manage the NooBaa operator and its resources
  description: |
    This plugin packages the entire NooBaa (aka Multi Cloud Gateway) CLI hence allows 
    you to manage the NooBaa operator and its resources.

    # Install the operator
    kubectl noobaa install -n <namespace>

    # Get the status of the operator
    kubectl noobaa status -n <namespace>

    It can also be used to create, delete, and manage NooBaa resources.
  platforms:
  - selector:
      matchLabels:
        os: linux
        arch: amd64
      uri: https://github.com/$GH_ORG/$GH_REPO/releases/download/$(get_noobaa_version 1)/noobaa-$(get_noobaa_version 1)-linux-amd64.tar.gz
      sha256: $(sha256sum $(generate_full_bin_path linux amd64) | cut -d ' ' -f 1)
      bin: noobaa-operator
  - selector:
      matchLabels:
        os: darwin
        arch: amd64
      uri: https://github.com/$GH_ORG/$GH_REPO/releases/download/$(get_noobaa_version 1)/noobaa-$(get_noobaa_version 1)-darwin-amd64.tar.gz
      sha256: $(sha256sum $(generate_full_bin_path darwin amd64) | cut -d ' ' -f 1)
      bin: noobaa-operator
  - selector:
      matchLabels:
        os: darwin
        arch: arm64
      uri: https://github.com/$GH_ORG/$GH_REPO/releases/download/$(get_noobaa_version 1)/noobaa-$(get_noobaa_version 1)-darwin-arm64.tar.gz
      sha256: $(sha256sum $(generate_full_bin_path darwin arm64) | cut -d ' ' -f 1)
      bin: noobaa-operator
  - selector:
      matchLabels:
        os: linux
        arch: arm64
      uri: https://github.com/$GH_ORG/$GH_REPO/releases/download/$(get_noobaa_version 1)/noobaa-$(get_noobaa_version 1)-linux-arm64.tar.gz
      sha256: $(sha256sum $(generate_full_bin_path linux arm64) | cut -d ' ' -f 1)
      bin: noobaa-operator
EOF
}

function generate_homebrew_formula() {
  cat <<EOF > $homebrew_path/noobaa.rb
class Noobaa < Formula
  desc "CLI for managing NooBaa S3 service on Kubernetes/Openshift"
  homepage "https://github.com/$GH_ORG/$GH_REPO"
  url "https://github.com/$GH_ORG/$GH_REPO.git",
      :tag      => "$(get_noobaa_version 1)",
      :revision => "$(git rev-list -n 1 $(get_noobaa_version 1) || echo HEAD)"
  head "https://github.com/$GH_ORG/$GH_REPO.git"

  depends_on "go" => [:build, :test]

  def install
    ENV.deparallelize # avoid parallel make jobs
    ENV["GOPATH"] = buildpath
    ENV["GO111MODULE"] = "on"
    ENV["GOPROXY"] = "https://proxy.golang.org"

    src = buildpath/"src/github.com/$GH_ORG/$GH_REPO"
    src.install buildpath.children
    src.cd do
      system "go", "mod", "vendor"
      system "go", "generate"
      system "go", "build"
      bin.install "noobaa-operator" => "noobaa"
    end
  end

  test do
    output = `#{bin}/noobaa version 2>&1`
    pos = output.index "CLI version: $(get_noobaa_version)"
    raise "Version check failed" if pos.nil?

    puts "Success"
  end
end
EOF
}

function generate_operator_bundle() {
  make gen-olm
  \cp -rf build/_output/olm/** $operator_bundle_path/
}

function main() {
  check_deps "${dependencies[@]}"
  build_cli_for_all
  build_oci ${OCI_ORG:-noobaa}
  generate_krew_manifest
  generate_homebrew_formula
  generate_operator_bundle
}

function usage() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  --help"
  echo "  --dry-run - enable dry run (will not prevent artifacts generation but will prevent git commits and pushes)"
  echo "  --gh-org <org> - set the github organization (default: noobaa)"
  echo "  --gh-repo <repo> - set the github repository (default: noobaa-operator)"
  echo "  --oci-org <org> - set the org name for the OCI image (default: noobaa)"
}

init "$@"

if [ -f "$dir/builder-pre.sh" ]; then
  DRY_RUN="$DRY_RUN" bash "$dir/builder-pre.sh"
fi

main "$@"

if [ -f "$dir/builder-post.sh" ]; then
  DRY_RUN="$DRY_RUN" bash "$dir/builder-post.sh"
fi
