#!/bin/sh
# Install the codeforge CLI from the latest GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/pointjungle/codeforge-cli/main/install.sh | sh
#
# Environment:
#   CODEFORGE_VERSION      a version to install (e.g. 0.1.0); default: the latest release
#   CODEFORGE_INSTALL_DIR  where to put the binary; default: $HOME/.local/bin
set -eu

REPO="https://github.com/pointjungle/codeforge-cli"
INSTALL_DIR="${CODEFORGE_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
	echo "codeforge install: $*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

need curl
need tar
need uname

case "$(uname -s)" in
Linux) os=linux ;;
Darwin) os=darwin ;;
*) fail "unsupported system $(uname -s); download a release from ${REPO}/releases instead" ;;
esac
case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) fail "unsupported architecture $(uname -m)" ;;
esac

version="${CODEFORGE_VERSION:-}"
if [ -z "$version" ]; then
	# /releases/latest redirects to /releases/tag/<tag>; no API call, no rate limit.
	latest=$(curl -fsSL -o /dev/null -w '%{url_effective}' "${REPO}/releases/latest") ||
		fail "could not find the latest release"
	tag="${latest##*/tag/}"
	[ "$tag" != "$latest" ] || tag=""
	[ -n "$tag" ] || fail "could not find the latest release"
	version="${tag#v}"
fi

archive="codeforge_${version}_${os}_${arch}.tar.gz"
base="${REPO}/releases/download/v${version}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading codeforge ${version} (${os}/${arch})…"
curl -fsSL "${base}/${archive}" -o "${tmp}/${archive}" || fail "download failed: ${base}/${archive}"
curl -fsSL "${base}/checksums.txt" -o "${tmp}/checksums.txt" || fail "download failed: checksums.txt"

expected=$(grep " ${archive}\$" "${tmp}/checksums.txt" | cut -d' ' -f1)
[ -n "$expected" ] || fail "no checksum for ${archive}"
if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "${tmp}/${archive}" | cut -d' ' -f1)
else
	actual=$(shasum -a 256 "${tmp}/${archive}" | cut -d' ' -f1)
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for ${archive}"

tar -xzf "${tmp}/${archive}" -C "$tmp" codeforge
mkdir -p "$INSTALL_DIR"
install -m 0755 "${tmp}/codeforge" "${INSTALL_DIR}/codeforge" 2>/dev/null ||
	{ cp "${tmp}/codeforge" "${INSTALL_DIR}/codeforge" && chmod 0755 "${INSTALL_DIR}/codeforge"; }

echo "Installed ${INSTALL_DIR}/codeforge"
case ":$PATH:" in
*":${INSTALL_DIR}:"*) ;;
*) echo "Add ${INSTALL_DIR} to your PATH to run codeforge." ;;
esac
