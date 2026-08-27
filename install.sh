#!/usr/bin/env bash
set -euo pipefail

REPO="parvosub/sylloge"
VERSION="${SYLLOGE_VERSION:-latest}"
INSTALL_DIR="${SYLLOGE_INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${SYLLOGE_CONFIG_DIR:-$HOME/.sylloge}"

# Resolve OS and architecture.
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ASSET="sylloge-${OS}-${ARCH}"
BASE="https://github.com/${REPO}/releases/${VERSION}"

echo "Downloading ${ASSET} from ${REPO}..."
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fSL -o "${TMPDIR}/${ASSET}" "${BASE}/download/${ASSET}"
curl -fSL -o "${TMPDIR}/${ASSET}.sha256" "${BASE}/download/${ASSET}.sha256"

# Verify checksum if present.
if [ -f "${TMPDIR}/${ASSET}.sha256" ]; then
  (cd "$TMPDIR" && shasum -a 256 -c "${ASSET}.sha256" >/dev/null)
  echo "Checksum verified."
fi

chmod +x "${TMPDIR}/${ASSET}"
mkdir -p "$INSTALL_DIR"
install -m 0755 "${TMPDIR}/${ASSET}" "${INSTALL_DIR}/sylloge"

# Write a default config if the user doesn't already have one.
mkdir -p "$CONFIG_DIR"
CONFIG_FILE="${CONFIG_DIR}/sylloge.toml"
if [ ! -f "$CONFIG_FILE" ]; then
  cat > "$CONFIG_FILE" <<'EOF'
# Sylloge configuration
[database]
path = "sylloge.db"

[llm]
provider = "local"
model = "qwen3:8b"
system_prompt = "You are a supportive teaching assistant writing a report-card comment for one student."

[api]
base_url = "http://localhost:11434/v1"
api_key = ""
EOF
  echo "Wrote default config to ${CONFIG_FILE}"
else
  echo "Config already exists at ${CONFIG_FILE}; leaving it unchanged."
fi

echo ""
echo "Sylloge installed to ${INSTALL_DIR}/sylloge"
echo ""
echo "To run:"
echo "  export SYLLOGE_CONFIG=${CONFIG_FILE}"
echo "  sylloge"
echo ""
echo "See https://github.com/${REPO} for configuration and Docker instructions."
