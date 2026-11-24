#!/bin/sh
# Install the PointJungle CLI from the latest GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/PointJungle/cli/main/install.sh | sh
#
# Environment:
#   PJ_VERSION      a version to install (e.g. 0.1.0); default: the latest release
#   PJ_INSTALL_DIR  where to put the binary; default: $HOME/.local/bin
set -eu

REPO="https://github.com/PointJungle/cli"
INSTALL_DIR="${PJ_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
	echo "pj install: $*" >&2
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

version="${PJ_VERSION:-}"
if [ -z "$version" ]; then
	# /releases/latest redirects to /releases/tag/<tag>; no API call, no rate limit.
	latest=$(curl -fsSL -o /dev/null -w '%{url_effective}' "${REPO}/releases/latest") ||
		fail "could not find the latest release"
	tag="${latest##*/tag/}"
	[ "$tag" != "$latest" ] || tag=""
	[ -n "$tag" ] || fail "could not find the latest release"
	version="${tag#v}"
fi

archive="pj_${version}_${os}_${arch}.tar.gz"
base="${REPO}/releases/download/v${version}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading pj ${version} (${os}/${arch})…"
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

tar -xzf "${tmp}/${archive}" -C "$tmp" pj
mkdir -p "$INSTALL_DIR"
install -m 0755 "${tmp}/pj" "${INSTALL_DIR}/pj" 2>/dev/null ||
	{ cp "${tmp}/pj" "${INSTALL_DIR}/pj" && chmod 0755 "${INSTALL_DIR}/pj"; }

echo "Installed ${INSTALL_DIR}/pj"

# Not on the PATH: print the command that adds it, for the user's default shell.
case ":$PATH:" in
*":${INSTALL_DIR}:"*) exit 0 ;;
esac

# Write the directory as $HOME/... when it is under the home directory.
case "$INSTALL_DIR" in
"$HOME"/*) dir_expr="\$HOME/${INSTALL_DIR#"$HOME"/}" ;;
*) dir_expr="$INSTALL_DIR" ;;
esac

shell_name=$(basename "${SHELL:-sh}")
case "$shell_name" in
zsh) rc="~/.zshrc" ;;
bash)
	if [ "$os" = darwin ]; then rc="~/.bash_profile"; else rc="~/.bashrc"; fi
	;;
fish) rc="" ;;
*) rc="~/.profile" ;;
esac

echo
echo "${INSTALL_DIR} is not on your PATH. To add it, run:"
echo
if [ "$shell_name" = fish ]; then
	echo "  fish_add_path ${dir_expr}"
else
	echo "  echo 'export PATH=\"${dir_expr}:\$PATH\"' >> ${rc}"
fi
echo
echo "then open a new terminal."
