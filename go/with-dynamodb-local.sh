#!/bin/bash

# Setup variables

TABLE_NAME="InterviewMockDepositsTable"
PK="depositId"
SK="itemType"
GSI="" # Add GSI names separated by space, e.g. "GSI1 GSI2"

# End setup variables

set -e

# DynamoDB Local settings
AWS_ENDPOINT_URL_DYNAMODB="http://localhost:8000"
DYNAMODB_PORT=8000
DYNAMODB_DIR="/tmp/DynamoDBLocal"
DYNAMODB_JAR="${DYNAMODB_DIR}/DynamoDBLocal.jar"
DYNAMODB_LIB="${DYNAMODB_DIR}/DynamoDBLocal_lib"
DYNAMODB_PID_FILE="/tmp/dynamodb.pid"
DYNAMODB_DB_PATH="$(mktemp -d -t dynamodb-local-db.XXXXXX)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
  echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if DynamoDB Local is running
is_dynamodb_running() {
  if [ -f "$DYNAMODB_PID_FILE" ]; then
    local pid=$(cat "$DYNAMODB_PID_FILE")
    if ps -p "$pid" > /dev/null 2>&1; then
      return 0
    else
      rm -f "$DYNAMODB_PID_FILE"
      return 1
    fi
  fi
  return 1
}

# Function to stop DynamoDB Local
stop_dynamodb() {
  if is_dynamodb_running; then
    local pid=$(cat "$DYNAMODB_PID_FILE")
    log_info "Stopping DynamoDB Local (PID: $pid)..."
    kill "$pid" || true
    rm -f "$DYNAMODB_PID_FILE"

    # Wait for process to stop
    local count=0
    while ps -p "$pid" > /dev/null 2>&1 && [ $count -lt 10 ]; do
      sleep 1
      count=$((count + 1))
    done

    if ps -p "$pid" > /dev/null 2>&1; then
      log_warn "Force killing DynamoDB Local..."
      kill -9 "$pid" || true
    fi
  fi

  # Kill any remaining DynamoDBLocal processes
  pkill -f "DynamoDBLocal.jar" > /dev/null 2>&1 || true

  log_info "DynamoDB Local stopped"
}

# Function to setup DynamoDB Local
setup_dynamodb() {
  log_info "Setting up DynamoDB Local..."

  # Check if dynamodb-local command exists in PATH
  if command -v dynamodb-local > /dev/null 2>&1; then
    log_info "Found dynamodb-local in PATH"
    return 0
  fi

  # Check if already downloaded
  if [ -f "$DYNAMODB_JAR" ]; then
    log_info "DynamoDB Local already exists at $DYNAMODB_JAR"
    return 0
  fi

  # Check if we're in CI environment
  if [ -n "$CI" ]; then
    log_info "CI environment detected, downloading DynamoDB Local..."
    mkdir -p "$DYNAMODB_DIR"
    curl -L https://d1ni2b6xgvw0s0.cloudfront.net/v2.x/dynamodb_local_latest.tar.gz | tar -xz -C "$DYNAMODB_DIR"
    if [ ! -f "$DYNAMODB_JAR" ]; then
      log_error "Failed to download DynamoDB Local"
      exit 1
    fi
    log_info "DynamoDB Local downloaded successfully"
    return 0
  fi

  # Not in CI and not found
  log_error "DynamoDB Local not found!"
  log_error "Please install it via Homebrew: brew install dynamodb-local"
  exit 1
}

# Function to start DynamoDB Local
start_dynamodb() {
  if is_dynamodb_running; then
    log_info "DynamoDB Local is already running"
    return 0
  fi

  log_info "Starting DynamoDB Local on port $DYNAMODB_PORT..."

  # Try using dynamodb-local command first
  if command -v dynamodb-local > /dev/null 2>&1; then
    nohup dynamodb-local -sharedDb -dbPath "$DYNAMODB_DB_PATH" -port "$DYNAMODB_PORT" > /dev/null 2>&1 &
    echo $! > "$DYNAMODB_PID_FILE"
  elif [ -f "$DYNAMODB_JAR" ]; then
    nohup java -Djava.library.path="$DYNAMODB_LIB" -jar "$DYNAMODB_JAR" -sharedDb -dbPath "$DYNAMODB_DB_PATH" -port "$DYNAMODB_PORT" > /dev/null 2>&1 &
    echo $! > "$DYNAMODB_PID_FILE"
  else
    log_error "Cannot start DynamoDB Local: neither command nor JAR file found"
    exit 1
  fi

  # Wait for DynamoDB to start
  log_info "Waiting for DynamoDB Local to start..."
  local count=0
  while ! curl -s "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null 2>&1 && [ $count -lt 30 ]; do
    sleep 1
    count=$((count + 1))
  done

  if [ $count -ge 30 ]; then
    log_error "DynamoDB Local failed to start within 30 seconds"
    stop_dynamodb
    exit 1
  fi

  log_info "DynamoDB Local started successfully"

  echo $TABLE_NAME
}

# Function to create table
create_table() {
  table=$1
  pk=$2
  sk=$3
  IFS=' ' read -ra GSI_ARRAY <<< "$4"

  log_info "Creating table: $table"

  # Check if table already exists
  if aws dynamodb describe-table --table-name "$table" --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null 2>&1; then
    log_info "Table $table already exists"
    return 0
  fi

  # Create table
  aws dynamodb create-table \
    --table-name "$table" \
    --key-schema \
        AttributeName="$pk",KeyType=HASH \
        AttributeName="$sk",KeyType=RANGE \
    --attribute-definitions \
        AttributeName="$pk",AttributeType=S \
        AttributeName="$sk",AttributeType=S \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --table-class STANDARD \
    --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null

  # Wait for table to be ready before adding GSIs
  aws dynamodb wait table-exists --table-name "$table" --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB"

  for index in "${GSI_ARRAY[@]}"; do
    log_info "\tGSI $index"
    aws dynamodb update-table \
      --table-name "$table" \
      --attribute-definitions AttributeName=$index,AttributeType=S AttributeName=$sk,AttributeType=S \
      --global-secondary-index-updates \
        '[{"Create":{"IndexName":"'$index'","KeySchema":[{"AttributeName":"'$index'","KeyType":"HASH"},{"AttributeName":"'$sk'","KeyType":"RANGE"}],"Projection":{"ProjectionType":"ALL"},"ProvisionedThroughput":{"ReadCapacityUnits":5,"WriteCapacityUnits":5}}}]' \
      --endpoint-url "$AWS_ENDPOINT_URL_DYNAMODB" > /dev/null
  done

  log_info "Table $table created successfully"
  echo -e
}

# Function to run the command
run_command() {
  if [ $# -eq 0 ]; then
    log_error "No command provided to run"
    log_error "Usage: $0 <command> [args...]"
    exit 1
  fi

  log_info "Running command: $*"

  # Run the command
  "$@"
}

# Signal handler for cleanup
cleanup() {
  log_info "Cleaning up..."
  stop_dynamodb
  if [ -n "$DYNAMODB_DB_PATH" ] && [ -d "$DYNAMODB_DB_PATH" ]; then
    rm -rf "$DYNAMODB_DB_PATH"
  fi
}

# Set up signal handlers
trap cleanup EXIT INT TERM

# Main execution
main() {
  setup_dynamodb
  start_dynamodb

  # create_table TABLE_NAME PK SK ["GSI1 GSI2..."]
  create_table $TABLE_NAME $PK $SK "$GSI"

  if [ $# -eq 0 ]; then
    log_info "No command provided. DynamoDB Local will continue running indefinitely..."
    # Keep process running indefinitely
    while true; do
      sleep 3600
    done
  else
    export TABLE_NAME="$TABLE_NAME"
    export AWS_ENDPOINT_URL_DYNAMODB="$AWS_ENDPOINT_URL_DYNAMODB"
    run_command "$@"
  fi
}

# Run main function with all arguments
main "$@"
