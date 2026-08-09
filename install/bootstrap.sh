#!/bin/sh

set -eu

readonly canonical_command='curl -fsSL https://get.stackdome.com/install.sh | sudo sh'

fail() {
    printf '%s\n' "stackdome installer: $*" >&2
    exit 1
}

cleanup() {
    if [ -n "${tmp_dir:-}" ] && [ -d "$tmp_dir" ]; then
        rm -rf "$tmp_dir"
    fi
}

if [ "$(id -u)" -ne 0 ]; then
    fail "root access required; run: $canonical_command"
fi

os=$(uname -s 2>/dev/null) || fail 'unable to determine operating system'
[ "$os" = Linux ] || fail "unsupported operating system: $os (Linux required)"

machine=$(uname -m 2>/dev/null) || fail 'unable to determine CPU architecture'
case "$machine" in
    x86_64|amd64)
        asset=stackdome-install-linux-amd64
        ;;
    aarch64|arm64)
        asset=stackdome-install-linux-arm64
        ;;
    *)
        fail "unsupported CPU architecture: $machine"
        ;;
esac

if command -v curl >/dev/null 2>&1; then
    download() {
        curl --fail --location --silent --show-error --output "$1" "$2"
    }
elif command -v wget >/dev/null 2>&1; then
    download() {
        wget -q -O "$1" "$2"
    }
else
    fail 'curl or wget is required'
fi

if command -v sha256sum >/dev/null 2>&1; then
    checksum() {
        sha256sum "$1"
    }
elif command -v shasum >/dev/null 2>&1; then
    checksum() {
        shasum -a 256 "$1"
    }
else
    fail 'sha256sum or shasum is required'
fi

release_tag=${STACKDOME_VERSION:-@STACKDOME_VERSION@}
[ -n "$release_tag" ] || fail 'STACKDOME_VERSION is required'
[ "$release_tag" != '@STACKDOME_VERSION@' ] || fail 'STACKDOME_VERSION is required'

release_base_url=${STACKDOME_RELEASE_BASE_URL:-https://github.com/Stackdome/stackdome/releases/download}

umask 077
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/stackdome-install.XXXXXX") || fail 'unable to create temporary directory'
trap cleanup 0
trap 'cleanup; exit 129' HUP
trap 'cleanup; exit 130' INT
trap 'cleanup; exit 143' TERM

checksums_file="$tmp_dir/checksums.txt"
installer_file="$tmp_dir/$asset"
release_url="$release_base_url/$release_tag"

download "$checksums_file" "$release_url/checksums.txt" || fail 'failed to download checksums.txt'

match_count=$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { count++ } END { print count + 0 }' "$checksums_file")
[ "$match_count" -eq 1 ] || fail "checksums.txt must contain exactly one entry for $asset"

expected=$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { print $1 }' "$checksums_file" | tr '[:upper:]' '[:lower:]')
case "$expected" in
    *[!0-9a-f]*|'') fail "invalid SHA-256 checksum for $asset" ;;
esac
[ "${#expected}" -eq 64 ] || fail "invalid SHA-256 checksum for $asset"

download "$installer_file" "$release_url/$asset" || fail "failed to download $asset"
checksum_line=$(checksum "$installer_file") || fail "failed to calculate checksum for $asset"
actual=$(printf '%s\n' "$checksum_line" | awk '{ print $1 }' | tr '[:upper:]' '[:lower:]')
[ "$actual" = "$expected" ] || fail "checksum mismatch for $asset"

if [ "${STACKDOME_DOWNLOAD_ONLY:-0}" = 1 ]; then
    printf '%s\n' "Verified $asset for $release_tag"
    exit 0
fi

needs_email=1
if [ "${1:-}" = upgrade ]; then
    needs_email=0
fi
for arg in "$@"; do
    case "$arg" in
        --email|--email=*) needs_email=0 ;;
    esac
done

if [ "$needs_email" -eq 1 ]; then
    if [ -r /dev/tty ] && [ -w /dev/tty ] && (: </dev/tty) 2>/dev/null; then
        printf 'Admin email: ' >/dev/tty
        IFS= read -r admin_email </dev/tty || fail 'unable to read admin email'
        [ -n "$admin_email" ] || fail 'admin email cannot be empty'
        set -- "$@" --email "$admin_email"
    else
        fail 'non-interactive install requires --email'
    fi
fi

chmod 0700 "$installer_file" || fail "failed to make $asset executable"
"$installer_file" "$@" || fail 'installer execution failed'
