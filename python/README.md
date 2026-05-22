# Interview Mock Deposits — Python

Python implementation of the interview exercise. CDK-deployed lambda service with DynamoDB-backed deposit handlers.

## Commands

```
./test.sh            # boots DynamoDB Local + runs pytest (uv)
./build.sh           # packages each lambda into build/<fn>/main.zip
cd cdk && npm install && npm run test:cdk    # run CDK stack tests (TS)
```

`./test.sh` calls `with-dynamodb-local.sh uv run pytest`, which starts DynamoDB Local, creates the application table, exports `AWS_ENDPOINT_URL_DYNAMODB` + `TABLE_NAME`, and runs the test suite. If you invoke `uv run pytest` directly, the integration tests skip with a clear message because that env var is missing.

## Layout

Mirrors `/go/src/` 1:1:

- `src/create_new_deposit/` — implemented reference. Handler + lambda-level integration tests.
- `src/deposit_in_provider/` — `pay_user` is a stub.
- `src/webhook/` — `handle_webhook` is a stub. Signature validation is already wired before the stub.
- `src/stream_consumer/` — implemented. `put_events` is intentionally log-only (mocked EventBridge publish).
- `src/partner_api/` — implemented. Consume; do not reimplement.
- `src/deposit/` — shared domain (`Deposit`, `Transaction`, `UserData`, `DepositStatus`) + DDB (de)serializers. `update.py` has stubs.
- `src/internal/`, `src/utils/` — shared helpers.

## Tools

- Python 3.12, `uv` for env/deps, `pytest` for tests, `ruff` for lint/format, `mypy --strict` for type checking.
