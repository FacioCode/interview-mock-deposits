#!/usr/bin/env bash
# Package each lambda into a self-contained zip.
#
# For each function:
#   1) install runtime deps (boto3 is excluded — provided by the Lambda runtime)
#      into build/<fn>/
#   2) copy the function's src/<fn>/ package directory (preserving its name)
#      into build/<fn>/, so Lambda's <fn>.handler.handle_request entry point
#      can resolve at cold start
#   3) copy the shared modules (deposit, partner_api, internal, utils) alongside
#      so absolute imports like `from deposit.create import …` resolve
#   4) zip the result as build/<fn>/main.zip

set -euo pipefail

cd "$(dirname "$0")"

functions=(
  create_new_deposit
  deposit_in_provider
  stream_consumer
  webhook
)

SHARED_MODULES=(deposit partner_api internal utils)

rm -rf build
mkdir -p build

# Runtime deps for the lambda zip. boto3 ships with the AWS Lambda Python
# runtime, so we don't bundle it.
RUNTIME_DEPS=()

package_function() {
  local fn=$1
  local target="build/${fn}"
  mkdir -p "${target}"

  if [ ${#RUNTIME_DEPS[@]} -gt 0 ]; then
    echo "Installing deps for ${fn}..."
    python3 -m pip install --quiet --target "${target}" "${RUNTIME_DEPS[@]}"
  fi

  cp -R "src/${fn}" "${target}/"
  for mod in "${SHARED_MODULES[@]}"; do
    cp -R "src/${mod}" "${target}/"
  done

  # Strip test files and Python caches from the production artifact.
  find "${target}" -type f -name "test_*.py" -delete
  find "${target}" -type d -name "__pycache__" -prune -exec rm -rf {} +

  (cd "${target}" && zip -qr "main.zip" . -x "main.zip")
  echo "Built ${target}/main.zip"
}

for fn in "${functions[@]}"; do
  package_function "${fn}"
done
