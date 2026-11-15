#!/bin/bash

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"
source "${SCRIPTS_DIR}/helpers-source.sh"

BIN_DIR=${REPO_ROOT}/bin
mkdir -p "${BIN_DIR}"

echo "${SCRIPT_NAME} is running... "

GORELEASER="$(tool_path goreleaser)"

"${GORELEASER}" healthcheck

APP=${APP_NAME}
# goreleaser ignores --output unless a single build is selected; .goreleaser.yml has `cli` and `bot`.
BUILD_ID=${BUILD_ID:-cli}

echo "Building ${APP} (goreleaser build id: ${BUILD_ID})..."

BIN_OUT="${BIN_DIR}/${APP}"

rm -rf "${BIN_OUT}"

"${GORELEASER}" build --skip=validate --clean --single-target --id "${BUILD_ID}" --output "${BIN_OUT}"

if [[ ! -f "${BIN_OUT}" ]]; then
  echo "Binary was not created at ${BIN_OUT}"
  exit 1
fi

echo "Binary compiled at ${BIN_OUT}"
