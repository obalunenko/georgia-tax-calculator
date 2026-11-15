#!/usr/bin/env bash

set -euo pipefail

required_variables=(
	SPECKIT_VERSION
	SPECKIT_COMPANION_VERSION
	SPECKIT_BROWNFIELD_VERSION
	SPECKIT_BUGFIX_VERSION
	SPECKIT_FEATURE_NUMBERING_VERSION
	SPECKIT_INTEGRATIONS
)
for variable_name in "${required_variables[@]}"; do
	if [[ -z "${!variable_name:-}" ]]; then
		printf 'Missing required environment variable: %s\n' "${variable_name}" >&2
		exit 1
	fi
done

readonly spec_kit_tag="v${SPECKIT_VERSION#v}"
readonly spec_kit_source="git+https://github.com/github/spec-kit.git@${spec_kit_tag}"
readonly companion_url="https://github.com/alfredoperez/speckit-companion/releases/download/speckit-ext-v${SPECKIT_COMPANION_VERSION}/companion-${SPECKIT_COMPANION_VERSION}.zip"
readonly brownfield_url="https://github.com/Quratulain-bilal/spec-kit-brownfield/archive/refs/tags/v${SPECKIT_BROWNFIELD_VERSION}.zip"
readonly bugfix_url="https://github.com/Quratulain-bilal/spec-kit-bugfix/archive/refs/tags/v${SPECKIT_BUGFIX_VERSION}.zip"

read -r -a integrations <<<"${SPECKIT_INTEGRATIONS}"
readonly integrations
# The first integration is the default; shared scripts and templates use its
# command spelling.
readonly default_integration="${integrations[0]}"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}" && git rev-parse --show-toplevel)"
readonly script_dir repo_root

cd "${repo_root}"

if ! command -v uv >/dev/null 2>&1; then
	printf 'uv is required. Install it from https://docs.astral.sh/uv/getting-started/installation/\n' >&2
	exit 1
fi

if [[ ! -f .speckit/feature-numbering/extension.yml ]]; then
	printf 'Missing local extension: .speckit/feature-numbering/extension.yml\n' >&2
	exit 1
fi

local_extension_version="$(awk '
	/^extension:/ { in_extension = 1; next }
	in_extension && /^  version:/ {
		gsub(/["'\''[:space:]]/, "", $2)
		print $2
		exit
	}
' .speckit/feature-numbering/extension.yml)"
readonly local_extension_version
if [[ "${local_extension_version}" != "${SPECKIT_FEATURE_NUMBERING_VERSION}" ]]; then
	printf 'Pinned feature-numbering version %s does not match local manifest version %s.\n' \
		"${SPECKIT_FEATURE_NUMBERING_VERSION}" "${local_extension_version:-missing}" >&2
	exit 1
fi

printf 'Installing GitHub Spec Kit %s...\n' "${spec_kit_tag}"
uv tool install --force specify-cli --from "${spec_kit_source}"

uv_bin_dir="$(uv tool dir --bin)"
specify_bin="${uv_bin_dir}/specify"
readonly uv_bin_dir specify_bin
if [[ ! -x "${specify_bin}" ]]; then
	printf 'Specify CLI was installed, but %s is not executable.\n' "${specify_bin}" >&2
	exit 1
fi

integration_options() {
	case "$1" in
	codex) printf '%s' '--skills' ;;
	esac
}

integration_installed() {
	python3 -c 'import json, sys; sys.exit(sys.argv[1] not in json.load(open(".specify/integration.json")).get("installed_integrations", []))' "$1"
}

if [[ ! -f .specify/integration.json ]]; then
	"${specify_bin}" init --here --force --non-interactive \
		--integration "${default_integration}" \
		--integration-options="$(integration_options "${default_integration}")" --script sh
fi

for integration in "${integrations[@]}"; do
	if integration_installed "${integration}"; then
		continue
	fi
	"${specify_bin}" integration install "${integration}" \
		--integration-options="$(integration_options "${integration}")" --script sh
done
"${specify_bin}" integration use "${default_integration}"

install_extension() {
	local extension_id="$1"
	local extension_url="$2"
	local priority="$3"

	# Spec Kit asks once for URL trust even with --force. These URLs are
	# explicitly pinned above; acknowledge only that prompt, then close stdin.
	printf 'y\n' | "${specify_bin}" extension add "${extension_id}" \
		--from "${extension_url}" \
		--force \
		--priority "${priority}"
}

install_extension companion "${companion_url}" 10
install_extension brownfield "${brownfield_url}" 10
install_extension bugfix "${bugfix_url}" 10
"${specify_bin}" extension add --dev --force --priority 5 .speckit/feature-numbering

# Spec Kit registers extension commands only for the default integration and
# adds them to another integration when it becomes the default.
for integration in "${integrations[@]}"; do
	if [[ "${integration}" != "${default_integration}" ]]; then
		"${specify_bin}" integration use "${integration}"
	fi
done
"${specify_bin}" integration use "${default_integration}"

# Rescaffolding rewrites shared scripts without their executable bit.
chmod +x .specify/scripts/bash/*.sh

printf '\nInstalled Spec Kit and project extensions:\n'
"${specify_bin}" version
"${specify_bin}" integration status
"${specify_bin}" extension list
printf '\nRestart your coding agent so it reloads generated skills (%s).\n' "${integrations[*]}"
