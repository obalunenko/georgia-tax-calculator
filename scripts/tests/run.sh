#!/bin/bash

set -eu pipefail

go version

go list -m

cat /usr/bin/log_build.txt

SCRIPT_NAME="$(basename "$0")"

echo "${SCRIPT_NAME} is running... "

GOTEST="go test -v "
if command -v "gotestsum" &>/dev/null; then
  GOTEST="gotestsum --format pkgname-and-test-fails --"
fi

${GOTEST} -race $(go list -m)/...

echo "${SCRIPT_NAME} done."
