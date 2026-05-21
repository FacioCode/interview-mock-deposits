#!/usr/bin/env bash
#
# postCreateCommand for the Codespaces / devcontainer setup.
# - Installs npm dependencies under go/
# - Pre-downloads the DynamoDB Local jar so the first test run is instant

set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> Installing Go subproject npm deps"
pushd go > /dev/null
npm install --no-audit --no-fund
popd > /dev/null

echo "==> Pre-downloading DynamoDB Local"
DYNAMODB_DIR="/tmp/DynamoDBLocal"
DYNAMODB_JAR="${DYNAMODB_DIR}/DynamoDBLocal.jar"
if [ ! -f "$DYNAMODB_JAR" ]; then
  mkdir -p "$DYNAMODB_DIR"
  curl -fsSL https://d1ni2b6xgvw0s0.cloudfront.net/v2.x/dynamodb_local_latest.tar.gz \
    | tar -xz -C "$DYNAMODB_DIR"
fi

echo "==> Setup complete. Run tests with:  (cd go && npm test)"
