#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

checkInstalled 'gofmt'

cd "${REPO_ROOT}"

echo "Running go fmt for all local packages"
echo "Making filelist"
FILES=($(find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./tools/vendor/*" -not -path "./.git/*"))


for f in "${FILES[@]}"; do
  echo "Fixing formatting at ${f}"
  gofmt -s -w "$f"
done

echo "${SCRIPT_NAME} done."
