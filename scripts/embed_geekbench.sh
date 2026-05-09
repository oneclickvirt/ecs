#!/usr/bin/env bash
# Populate cpu/geekbench_bins in the cputest module cache with the
# platform-specific Geekbench binary before each goreleaser cross-compile pass.
#
# Called by .goreleaser.yaml universal build pre-hook.
# Requires GOOS and GOARCH to be set in the environment (goreleaser injects them).
#
# Mirrors the per-target logic in oneclickvirt/cputest .github/workflows/main.yaml:
#   linux/amd64  → Geekbench 6 (x86_64)
#   linux/arm64  → Geekbench 6 ARM Preview (aarch64)
#   linux/arm    → Geekbench 5 ARM Preview (better glibc compat on 32-bit ARM)
#   other        → no embed; runtime falls back to geekbench in PATH

GOOS="${GOOS:-}"
GOARCH="${GOARCH:-}"
GB6_VER="6.7.1"
GB5_VER="5.5.1"
BASE="https://cdn.geekbench.com"

CPUTEST_VERSION=$(grep "github.com/oneclickvirt/cputest" go.mod | awk '{print $2}')
MODCACHE=$(go env GOMODCACHE)
CPUTEST_DIR="${MODCACHE}/github.com/oneclickvirt/cputest@${CPUTEST_VERSION}"
BINS="${CPUTEST_DIR}/cpu/geekbench_bins"

echo "[geekbench-embed] target=${GOOS}/${GOARCH}  bins=${BINS}"

# Module cache entries are read-only by default; make writable before editing.
chmod -R u+w "${CPUTEST_DIR}" 2>/dev/null || true
mkdir -p "${BINS}"

# Always clear previous platform's files first to avoid stale cross-platform data.
rm -f "${BINS}/geekbench" \
      "${BINS}/geekbench_x86_64" \
      "${BINS}/geekbench_aarch64" \
      "${BINS}/geekbench_armv7" \
      "${BINS}/geekbench.plar"

_download() {
    local url="$1" out="$2"
    wget -q "${url}" -O "${out}" 2>/dev/null \
        || curl -fsSL "${url}" -o "${out}"
}

case "${GOOS}/${GOARCH}" in

    linux/amd64)
        cd /tmp
        if ! _download "${BASE}/Geekbench-${GB6_VER}-Linux.tar.gz" gb.tar.gz; then
            echo "[geekbench-embed] WARNING: download failed; no embedded Geekbench for linux/amd64."
            exit 0
        fi
        tar -xf gb.tar.gz
        cp "Geekbench-${GB6_VER}-Linux/geekbench6"       "${BINS}/geekbench"
        cp "Geekbench-${GB6_VER}-Linux/geekbench_x86_64" "${BINS}/"
        cp "Geekbench-${GB6_VER}-Linux/geekbench.plar"   "${BINS}/"
        chmod +x "${BINS}/geekbench" "${BINS}/geekbench_x86_64"
        rm -rf "Geekbench-${GB6_VER}-Linux" gb.tar.gz
        echo "[geekbench-embed] Geekbench 6 (linux/amd64) embedded."
        ;;

    linux/arm64)
        cd /tmp
        if ! _download "${BASE}/Geekbench-${GB6_VER}-LinuxARMPreview.tar.gz" gb.tar.gz; then
            echo "[geekbench-embed] WARNING: download failed; no embedded Geekbench for linux/arm64."
            exit 0
        fi
        tar -xf gb.tar.gz
        cp "Geekbench-${GB6_VER}-LinuxARMPreview/geekbench6"        "${BINS}/geekbench"
        cp "Geekbench-${GB6_VER}-LinuxARMPreview/geekbench_aarch64" "${BINS}/"
        cp "Geekbench-${GB6_VER}-LinuxARMPreview/geekbench_armv7"   "${BINS}/"
        cp "Geekbench-${GB6_VER}-LinuxARMPreview/geekbench.plar"    "${BINS}/"
        chmod +x "${BINS}/geekbench"
        rm -rf "Geekbench-${GB6_VER}-LinuxARMPreview" gb.tar.gz
        echo "[geekbench-embed] Geekbench 6 ARM Preview (linux/arm64) embedded."
        ;;

    linux/arm)
        # GB5 has better glibc compatibility on 32-bit ARM than GB6.
        cd /tmp
        if ! _download "${BASE}/Geekbench-${GB5_VER}-LinuxARMPreview.tar.gz" gb.tar.gz; then
            echo "[geekbench-embed] WARNING: download failed; no embedded Geekbench for linux/arm."
            exit 0
        fi
        tar -xf gb.tar.gz
        cp "Geekbench-${GB5_VER}-LinuxARMPreview/geekbench5"        "${BINS}/geekbench"
        cp "Geekbench-${GB5_VER}-LinuxARMPreview/geekbench_aarch64" "${BINS}/"
        cp "Geekbench-${GB5_VER}-LinuxARMPreview/geekbench_armv7"   "${BINS}/"
        cp "Geekbench-${GB5_VER}-LinuxARMPreview/geekbench.plar"    "${BINS}/"
        chmod +x "${BINS}/geekbench"
        rm -rf "Geekbench-${GB5_VER}-LinuxARMPreview" gb.tar.gz
        echo "[geekbench-embed] Geekbench 5 ARM Preview (linux/arm) embedded."
        ;;

    *)
        echo "[geekbench-embed] No Geekbench binary for ${GOOS}/${GOARCH}; runtime PATH fallback applies."
        ;;
esac

echo "[geekbench-embed] --- bins contents ---"
ls -lh "${BINS}/" || true
