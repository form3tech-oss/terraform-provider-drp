#!/bin/bash

set -e

VERSION=$(echo ${TRAVIS_TAG} | sed 's/v//g')

BINARY="bin/darwin/amd64/terraform-provider-drp"
BINARY_TAGGED=${BINARY}_${TRAVIS_TAG}
mv ${BINARY} ${BINARY_TAGGED}
zip -j "${BINARY}_${VERSION}_darwin_amd64.zip" "${BINARY_TAGGED}"

BINARY="bin/linux/amd64/terraform-provider-drp"
BINARY_TAGGED=${BINARY}_${TRAVIS_TAG}
mv ${BINARY} ${BINARY_TAGGED}
zip -j "${BINARY}_${VERSION}_linux_amd64.zip" "${BINARY_TAGGED}"
