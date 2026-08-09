#!/bin/sh

set -eu

readonly canonical_command='curl -fsSL https://get.stackdome.com/install.sh | sudo sh'

output_value=
expect_output_value=0
for arg in "$@"; do
    if [ "$expect_output_value" -eq 1 ]; then
        output_value=$arg
        expect_output_value=0
        continue
    fi
    case "$arg" in
        --output) expect_output_value=1 ;;
        --output=*) output_value=${arg#--output=} ;;
    esac
done
json_output=0
[ "$output_value" = json ] && json_output=1

fail() {
    detail=$1
    public=${2:-transport failed}
    printf '%s\n' "stackdome installer: $detail" >&2
    if [ "$json_output" -eq 1 ]; then
        printf '{"status":"error","phase":"transport","message":"%s"}\n' "$public"
    fi
    exit 1
}

cleanup() {
    if [ -n "${tmp_dir:-}" ] && [ -d "$tmp_dir" ]; then
        rm -rf "$tmp_dir"
    fi
}

if [ "$(id -u)" -ne 0 ]; then
    fail "root access required; run: $canonical_command" 'root access required'
fi

os=$(uname -s 2>/dev/null) || fail 'unable to determine operating system' 'unsupported platform'
[ "$os" = Linux ] || fail "unsupported operating system: $os (Linux required)" 'unsupported platform'

machine=$(uname -m 2>/dev/null) || fail 'unable to determine CPU architecture' 'unsupported platform'
case "$machine" in
    x86_64|amd64)
        asset=stackdome-install-linux-amd64
        ;;
    aarch64|arm64)
        asset=stackdome-install-linux-arm64
        ;;
    *)
        fail "unsupported CPU architecture: $machine" 'unsupported platform'
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
    fail 'curl or wget is required' 'download tool unavailable'
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
    fail 'sha256sum or shasum is required' 'checksum tool unavailable'
fi

release_tag=${STACKDOME_VERSION:-@STACKDOME_VERSION@}
[ -n "$release_tag" ] || fail 'STACKDOME_VERSION is required' 'release version unavailable'
[ "$release_tag" != '@STACKDOME_VERSION@' ] || fail 'STACKDOME_VERSION is required' 'release version unavailable'

release_base_url=${STACKDOME_RELEASE_BASE_URL:-https://github.com/Stackdome/stackdome/releases/download}

umask 077
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/stackdome-install.XXXXXX") || fail 'unable to create temporary directory' 'temporary directory creation failed'
trap cleanup 0
trap 'cleanup; exit 129' HUP
trap 'cleanup; exit 130' INT
trap 'cleanup; exit 143' TERM

checksums_file="$tmp_dir/checksums.txt"
installer_file="$tmp_dir/$asset"
release_url="$release_base_url/$release_tag"

download "$checksums_file" "$release_url/checksums.txt" || fail 'failed to download checksums.txt' 'checksum manifest download failed'

match_count=$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { count++ } END { print count + 0 }' "$checksums_file")
[ "$match_count" -eq 1 ] || fail "checksums.txt must contain exactly one entry for $asset" 'checksum manifest invalid'

expected=$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { print $1 }' "$checksums_file" | tr '[:upper:]' '[:lower:]')
case "$expected" in
    *[!0-9a-f]*|'') fail "invalid SHA-256 checksum for $asset" 'checksum manifest invalid' ;;
esac
[ "${#expected}" -eq 64 ] || fail "invalid SHA-256 checksum for $asset" 'checksum manifest invalid'

download "$installer_file" "$release_url/$asset" || fail "failed to download $asset" 'installer download failed'
checksum_line=$(checksum "$installer_file") || fail "failed to calculate checksum for $asset" 'installer checksum verification failed'
actual=$(printf '%s\n' "$checksum_line" | awk '{ print $1 }' | tr '[:upper:]' '[:lower:]')
[ "$actual" = "$expected" ] || fail "checksum mismatch for $asset" 'installer checksum verification failed'

if [ "${STACKDOME_DOWNLOAD_ONLY:-0}" = 1 ]; then
    if [ "$json_output" -eq 1 ]; then
        printf '{"status":"verified"}\n'
    else
        printf '%s\n' "Verified $asset for $release_tag"
    fi
    exit 0
fi

needs_email=1
needs_domain=1
if [ "${1:-}" = upgrade ]; then
    needs_email=0
    needs_domain=0
fi
for arg in "$@"; do
    case "$arg" in
        --email|--email=*) needs_email=0 ;;
        --domain|--domain=*) needs_domain=0 ;;
    esac
done

interactive=0
if [ "$json_output" -eq 0 ] && [ -r /dev/tty ] && [ -w /dev/tty ] && (: </dev/tty) 2>/dev/null; then
    interactive=1
fi

if [ "$needs_email" -eq 1 ]; then
    if [ "$json_output" -eq 1 ]; then
        fail 'JSON install requires --email' 'non-interactive install requires --email'
    elif [ "$interactive" -eq 1 ]; then
        printf 'Admin email: ' >/dev/tty
        IFS= read -r admin_email </dev/tty || fail 'unable to read admin email' 'admin email is required'
        [ -n "$admin_email" ] || fail 'admin email cannot be empty' 'admin email is required'
        set -- "$@" --email "$admin_email"
    else
        fail 'non-interactive install requires --email' 'non-interactive install requires --email'
    fi
fi

if [ "$needs_domain" -eq 1 ] && [ "$interactive" -eq 1 ]; then
    printf 'Custom domain (press Enter to use automatic nip.io): ' >/dev/tty
    IFS= read -r custom_domain </dev/tty || fail 'unable to read custom domain' 'domain input failed'
    if [ -n "$custom_domain" ]; then
        set -- "$@" --domain "$custom_domain"
    fi
fi

chmod 0700 "$installer_file" || fail "failed to make $asset executable" 'installer preparation failed'
installer_stdout="$tmp_dir/installer.stdout"
installer_status=0
"$installer_file" "$@" >"$installer_stdout" || installer_status=$?
if ! cat "$installer_stdout"; then
    printf '%s\n' 'stackdome installer: failed to forward installer output' >&2
    exit 1
fi
if [ "$installer_status" -ne 0 ]; then
    if [ "$json_output" -eq 1 ] && [ ! -s "$installer_stdout" ]; then
        printf '{"status":"error","phase":"transport","message":"installer execution failed"}\n'
    fi
    printf '%s\n' 'stackdome installer: installer execution failed' >&2
    exit "$installer_status"
fi
