#!/usr/bin/env bash
# Canonical app-test entrypoint for the Python port.
# Boots DynamoDB Local, exports AWS_ENDPOINT_URL_DYNAMODB + TABLE_NAME,
# then runs pytest under uv.

set -euo pipefail

cd "$(dirname "$0")"

exec ./with-dynamodb-local.sh uv run pytest "$@"
