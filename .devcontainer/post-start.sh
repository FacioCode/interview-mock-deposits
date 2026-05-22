#!/usr/bin/env bash
#
# postStartCommand for Codespaces / devcontainer.
# Runs on every container start. Ensures DynamoDB Local is up on
# :8000 and the application table exists, so `go test ./...` from
# the IDE works the same way as `npm test` from the terminal.

set -euo pipefail

# DynamoDB Local doesn't validate region/credentials, but the AWS CLI refuses
# to issue a request without them. Provide dummies if the environment hasn't.
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-local}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-local}"

DYNAMODB_DIR="/tmp/DynamoDBLocal"
DYNAMODB_JAR="${DYNAMODB_DIR}/DynamoDBLocal.jar"
DYNAMODB_LIB="${DYNAMODB_DIR}/DynamoDBLocal_lib"
DYNAMODB_PORT=8000
AWS_ENDPOINT_URL_DYNAMODB="http://localhost:${DYNAMODB_PORT}"
TABLE_NAME="${TABLE_NAME:-InterviewMockDepositsTable}"

is_running() {
  curl -s -m 1 "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null 2>&1
}

# --- start DDB Local if it's not already up -----------------------------
if ! is_running; then
  if [ ! -f "$DYNAMODB_JAR" ]; then
    echo "DDB Local jar not found at $DYNAMODB_JAR — was postCreate skipped?" >&2
    exit 1
  fi

  DBPATH="/tmp/dynamodb-local-data"
  mkdir -p "$DBPATH"
  nohup java -Djava.library.path="$DYNAMODB_LIB" \
    -jar "$DYNAMODB_JAR" \
    -sharedDb -dbPath "$DBPATH" -port "$DYNAMODB_PORT" \
    >> /tmp/dynamodb.log 2>&1 &

  for _ in {1..30}; do
    if is_running; then break; fi
    sleep 1
  done
  is_running || { echo "DynamoDB Local failed to start within 30s" >&2; exit 1; }
  echo "==> DynamoDB Local started on :${DYNAMODB_PORT}"
fi

# --- create the application table if it's missing -----------------------
if ! aws dynamodb describe-table \
      --table-name "$TABLE_NAME" \
      --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null 2>&1; then
  echo "==> Creating table $TABLE_NAME"
  aws dynamodb create-table \
    --table-name "$TABLE_NAME" \
    --key-schema AttributeName=depositId,KeyType=HASH \
    --attribute-definitions \
        AttributeName=depositId,AttributeType=S \
        AttributeName=userId,AttributeType=S \
    --global-secondary-indexes \
        '[{"IndexName":"userIndex","KeySchema":[{"AttributeName":"userId","KeyType":"HASH"}],"Projection":{"ProjectionType":"ALL"},"ProvisionedThroughput":{"ReadCapacityUnits":5,"WriteCapacityUnits":5}}]' \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null
  aws dynamodb wait table-exists \
    --table-name "$TABLE_NAME" \
    --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB"
  aws dynamodb update-time-to-live \
    --table-name "$TABLE_NAME" \
    --time-to-live-specification "Enabled=true, AttributeName=ttl" \
    --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null
  echo "==> Table $TABLE_NAME ready"
fi
