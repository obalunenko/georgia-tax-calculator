#!/usr/bin/env bash

set -euo pipefail

readonly CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(dirname "${CURRENT_DIR}")"
readonly GO_MOD_FILE="${ROOT_DIR}/go.mod"

function usage() {
  cat <<EOF
Usage: $(basename "$0") <go-version>

Examples:
  $(basename "$0") 1.26
  $(basename "$0") 1.26.2
EOF
}

function validateGoVersion() {
  local goVersion="${1}"
  if [[ ! "${goVersion}" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
    echo "error: invalid Go version '${goVersion}'. Expected <major>.<minor> or <major>.<minor>.<patch>." >&2
    usage
    exit 1
  fi
}

function toCIVersion() {
  local goVersion="${1}"
  if [[ "${goVersion}" =~ ^([0-9]+\.[0-9]+)(\.[0-9]+)?$ ]]; then
    echo "${BASH_REMATCH[1]}.x"
    return 0
  fi

  echo "error: unable to derive CI version from '${goVersion}'." >&2
  exit 1
}

function rewriteFile() {
  local file="${1}"
  shift

  local tmpFile="${file}.tmp"
  sed -E "$@" "${file}" > "${tmpFile}"

  if cmp -s "${file}" "${tmpFile}"; then
    rm -f "${tmpFile}"
    return 0
  fi

  local mode=""
  if mode="$(stat -f '%Lp' "${file}" 2>/dev/null)"; then
    chmod "${mode}" "${tmpFile}"
  elif mode="$(stat -c '%a' "${file}" 2>/dev/null)"; then
    chmod "${mode}" "${tmpFile}"
  fi

  mv "${tmpFile}" "${file}"
}

# Replace go/toolchain version directives in all go.mod files.
function bumpModFiles() {
  local goVersion="${1}"

  while IFS= read -r -d '' modFile; do
    rewriteFile "${modFile}" \
      -e "s/^go [0-9]+\.[0-9]+(\.[0-9]+)?$/go ${goVersion}/g" \
      -e "s/^toolchain go[0-9]+\.[0-9]+(\.[0-9]+)?$/toolchain go${goVersion}/g"
  done < <(find "${ROOT_DIR}" -name "go.mod" -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*" -print0)
}

# Replace Go version in Makefiles.
function bumpMakeFiles() {
  local goVersion="${1}"

  while IFS= read -r -d '' makeFile; do
    rewriteFile "${makeFile}" \
      -e "s/^(GOVERSION[[:space:]]*[:?]?=[[:space:]]*)[0-9]+\.[0-9]+(\.[0-9]+)?/\1${goVersion}/g"
  done < <(find "${ROOT_DIR}" -type f \( -name "Makefile" -o -name "*.mk" \) -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*" -print0)
}

# Replace matrix go-version in GitHub Actions workflows.
function bumpCIMatrix() {
  local ciGoVersion="${1}"

  local workflowsDir="${ROOT_DIR}/.github/workflows"
  if [[ ! -d "${workflowsDir}" ]]; then
    return 0
  fi

  while IFS= read -r -d '' workflowFile; do
    rewriteFile "${workflowFile}" \
      -e "s/(go-version:[[:space:]]*\\[)[^]]+(\\])/\\1${ciGoVersion}\\2/g" \
      -e "s/^([[:space:]]*go-version:[[:space:]]*)[0-9]+\.[0-9]+(\.[0-9]+)?(\.x)?$/\\1${ciGoVersion}/g"
  done < <(find "${workflowsDir}" -type f \( -name "*.yml" -o -name "*.yaml" \) -print0)
}

# Replace golang image tags, e.g. golang:1.25 or golang:1.25.3.
function bumpGolangDockerImages() {
  local goVersion="${1}"

  while IFS= read -r -d '' file; do
    rewriteFile "${file}" \
      -e "s/(golang:)[0-9]+\.[0-9]+(\.[0-9]+)?/\\1${goVersion}/g"
  done < <(find "${ROOT_DIR}" -type f \( -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "Dockerfile*" \) -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*" -print0)
}

function extractCurrentVersion() {
  grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?$' "${GO_MOD_FILE}" | awk '{print $2}' | head -n 1
}

function main() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 1
  fi

  local goVersion="${1}"
  validateGoVersion "${goVersion}"

  local currentGoVersion
  currentGoVersion="$(extractCurrentVersion)"
  local ciGoVersion
  ciGoVersion="$(toCIVersion "${goVersion}")"

  echo "Updating Go version:"
  echo " - Current: ${currentGoVersion}"
  echo " - New (go.mod/Makefile/images): ${goVersion}"
  echo " - New (GitHub Actions matrix): ${ciGoVersion}"

  bumpModFiles "${goVersion}"
  bumpMakeFiles "${goVersion}"
  bumpGolangDockerImages "${goVersion}"
  bumpCIMatrix "${ciGoVersion}"
}

main "$@"
