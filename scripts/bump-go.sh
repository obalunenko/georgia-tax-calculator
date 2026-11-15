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

function validate_go_version() {
  local go_version="${1}"
  if [[ ! "${go_version}" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
    echo "error: invalid Go version '${go_version}'. Expected <major>.<minor> or <major>.<minor>.<patch>." >&2
    usage
    exit 1
  fi
}

function to_ci_version() {
  local go_version="${1}"
  if [[ "${go_version}" =~ ^([0-9]+\.[0-9]+)(\.[0-9]+)?$ ]]; then
    echo "${BASH_REMATCH[1]}.x"
    return 0
  fi

  echo "error: unable to derive CI version from '${go_version}'." >&2
  exit 1
}

function rewrite_file() {
  local file="${1}"
  shift

  local tmp_file="${file}.tmp"
  sed -E "$@" "${file}" > "${tmp_file}"

  if cmp -s "${file}" "${tmp_file}"; then
    rm -f "${tmp_file}"
    return 0
  fi

  local mode=""
  if mode="$(stat -f '%Lp' "${file}" 2>/dev/null)"; then
    chmod "${mode}" "${tmp_file}"
  elif mode="$(stat -c '%a' "${file}" 2>/dev/null)"; then
    chmod "${mode}" "${tmp_file}"
  fi

  mv "${tmp_file}" "${file}"
}

# Replace go/toolchain version directives in all go.mod files.
function bump_mod_files() {
  local go_version="${1}"

  while IFS= read -r -d '' mod_file; do
    rewrite_file "${mod_file}" \
      -e "s/^go [0-9]+\.[0-9]+(\.[0-9]+)?$/go ${go_version}/g" \
      -e "s/^toolchain go[0-9]+\.[0-9]+(\.[0-9]+)?$/toolchain go${go_version}/g"
  done < <(find "${ROOT_DIR}" -name "go.mod" -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*" -print0)
}

# Replace Go version in Makefiles.
function bump_make_files() {
  local go_version="${1}"

  while IFS= read -r -d '' make_file; do
    rewrite_file "${make_file}" \
      -e "s/^(GOVERSION[[:space:]]*[:?]?=[[:space:]]*)[0-9]+\.[0-9]+(\.[0-9]+)?/\1${go_version}/g"
  done < <(find "${ROOT_DIR}" -type f \( -name "Makefile" -o -name "*.mk" \) -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*" -print0)
}

# Replace matrix go-version in GitHub Actions workflows.
function bump_ci_matrix() {
  local ci_go_version="${1}"

  local workflows_dir="${ROOT_DIR}/.github/workflows"
  if [[ ! -d "${workflows_dir}" ]]; then
    return 0
  fi

  while IFS= read -r -d '' workflow_file; do
    rewrite_file "${workflow_file}" \
      -e "s/(go-version:[[:space:]]*\\[)[^]]+(\\])/\\1${ci_go_version}\\2/g" \
      -e "s/^([[:space:]]*go-version:[[:space:]]*)[0-9]+\.[0-9]+(\.[0-9]+)?(\.x)?$/\\1${ci_go_version}/g"
  done < <(find "${workflows_dir}" -type f \( -name "*.yml" -o -name "*.yaml" \) -print0)
}

# Replace golang image tags, e.g. golang:1.25 or golang:1.25.3.
function bump_golang_docker_images() {
  local go_version="${1}"

  while IFS= read -r -d '' file; do
    rewrite_file "${file}" \
      -e "s/(golang:)[0-9]+\.[0-9]+(\.[0-9]+)?/\\1${go_version}/g"
  done < <(find "${ROOT_DIR}" -type f \( -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "Dockerfile*" \) -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*" -print0)
}

function extract_current_version() {
  grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?$' "${GO_MOD_FILE}" | awk '{print $2}' | head -n 1
}

function main() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 1
  fi

  local go_version="${1}"
  validate_go_version "${go_version}"

  local current_go_version
  current_go_version="$(extract_current_version)"
  local ci_go_version
  ci_go_version="$(to_ci_version "${go_version}")"

  echo "Updating Go version:"
  echo " - Current: ${current_go_version}"
  echo " - New (go.mod/Makefile/images): ${go_version}"
  echo " - New (GitHub Actions matrix): ${ci_go_version}"

  bump_mod_files "${go_version}"
  bump_make_files "${go_version}"
  bump_golang_docker_images "${go_version}"
  bump_ci_matrix "${ci_go_version}"
}

main "$@"
