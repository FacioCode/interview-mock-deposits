#!/usr/bin/env bash
#
# postCreateCommand for the Codespaces / devcontainer setup.
# - Installs npm dependencies for each language folder present
# - Installs uv (Python package manager) if python/ exists
# - Pre-downloads the DynamoDB Local jar so the first test run is instant

set -euo pipefail

cd "$(dirname "$0")/.."

if [ -f go/package.json ]; then
  echo "==> Installing Go subproject npm deps"
  pushd go > /dev/null
  npm install --no-audit --no-fund
  popd > /dev/null
fi

if [ -f python/package.json ]; then
  echo "==> Installing Python subproject npm deps"
  pushd python > /dev/null
  npm install --no-audit --no-fund
  popd > /dev/null
fi

if [ -d python ] && ! command -v uv > /dev/null 2>&1; then
  echo "==> Installing uv"
  curl -LsSf https://astral.sh/uv/install.sh | sh
  # uv installs to ~/.local/bin; export for the rest of this script
  export PATH="$HOME/.local/bin:$PATH"
fi

if [ -d python ]; then
  echo "==> Syncing Python deps with uv"
  pushd python > /dev/null
  uv sync --extra dev
  popd > /dev/null
fi

echo "==> Pre-downloading DynamoDB Local"
DYNAMODB_DIR="/tmp/DynamoDBLocal"
DYNAMODB_JAR="${DYNAMODB_DIR}/DynamoDBLocal.jar"
if [ ! -f "$DYNAMODB_JAR" ]; then
  mkdir -p "$DYNAMODB_DIR"
  curl -fsSL https://d1ni2b6xgvw0s0.cloudfront.net/v2.x/dynamodb_local_latest.tar.gz \
    | tar -xz -C "$DYNAMODB_DIR"
fi

echo "==> Setup complete. Run tests with:"
[ -d go ] && echo "    (cd go && ./test.sh)"
[ -d python ] && echo "    (cd python && ./test.sh)"
