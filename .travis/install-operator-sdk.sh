#!/bin/bash
SDK_VERSION="v0.16.0"
SDK_RELEASE="https://github.com/operator-framework/operator-sdk/releases/download/${SDK_VERSION}/operator-sdk-${SDK_VERSION}-x86_64-linux-gnu"

echo "install-sdk - installing version ${SDK_VERSION}"
curl "${SDK_RELEASE}" -Lo $GOPATH/bin/operator-sdk
chmod +x $GOPATH/bin/operator-sdk
$GOPATH/bin/operator-sdk version
