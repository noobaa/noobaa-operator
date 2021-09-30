#!/bin/bash

if [ -z "${OPERATOR_SDK_VERSION}" ]; then
    echo "OPERATOR_SDK_VERSION should be defined as an env variable (usually in makefile)"
    exit 1
fi

if [ -z "${OPERATOR_SDK}" ]; then
    echo "OPERATOR_SDK should be defined as an env variable (usually in makefile)"
    exit 1
fi

mkdir -p "$(dirname "${OPERATOR_SDK}")"
if [ -x "${OPERATOR_SDK}" ]
then
    "${OPERATOR_SDK}" version | grep -q "\"${OPERATOR_SDK_VERSION}\""
    if [ $? -eq 0 ]
    then
        exit 0
    fi
fi

PLATFORM="$(uname)"
ARCH=$(uname -m)
ARCH_TAG=""
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH_TAG="x86_64"
elif [[ "$ARCH" == "aarch64" ]]; then
    # NOTE: newer version of operator-sdk use 'arm64' tag
    ARCH_TAG="aarch64"
elif [[ "$ARCH" == "arm64" && "${PLATFORM}" == "Darwin" ]]; then
    # Current in-use operator-sdk version does not support arm64 for MacOS
    ARCH_TAG="x86_64"    
else
    echo "unsupported: $ARCH on ${PLATFORM}"
    exit 1
fi

if [ "${PLATFORM}" == "Darwin" ] 
then 
    SDK_RELEASE="https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk-${OPERATOR_SDK_VERSION}-${ARCH_TAG}-apple-darwin"
else 
    # Assuming that if not darwin then running on linux
    SDK_RELEASE="https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk-${OPERATOR_SDK_VERSION}-${ARCH_TAG}-linux-gnu"
fi

echo "installing version ${OPERATOR_SDK_VERSION}"
curl -f "${SDK_RELEASE}" -Lo "${OPERATOR_SDK}"
if [ $? -ne 0 ]
then
    echo "could not download and install ${SDK_RELEASE}"
    exit 1
fi
chmod +x "${OPERATOR_SDK}"
