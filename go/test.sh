#!/usr/bin/env bash
# Canonical app-test entrypoint for the Go port.
# Boots DynamoDB Local, exports AWS_ENDPOINT_URL_DYNAMODB + TABLE_NAME,
# then runs `go test` over ./src/....

set -euo pipefail

cd "$(dirname "$0")"

exec ./with-dynamodb-local.sh go test -p 1 -coverprofile=coverage.txt -covermode=atomic ./src/... "$@"
