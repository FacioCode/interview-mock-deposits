"""Integration tests for the create_new_deposit lambda.

Drive handle_request end-to-end against DynamoDB Local. Run via:

    ./test.sh

The wrapper starts DynamoDB Local, creates the table, and exports
AWS_ENDPOINT_URL_DYNAMODB so boto3 points at it. These tests skip with a
clear message when that env var is not set.
"""
import os

import pytest

from create_new_deposit.handler import handle_request
from deposit import get_deposit_by_user_id
from internal.events import (
    PENDING_TRANSACTION_SOURCE,
    PENDING_TRANSACTION_TYPE,
)
from internal.test_utils import generate_random_id


def _skip_if_no_dynamo_local() -> None:
    if not os.environ.get("AWS_ENDPOINT_URL_DYNAMODB"):
        pytest.skip("DynamoDB Local not detected; run via `./test.sh`")


def _make_event_detail(customer_id: str, transaction_id: str) -> dict[str, str | dict[str, str]]:
    return {
        "transactionId": transaction_id,
        "customerId": customer_id,
        "amount": "100.00",
        "name": "Test User",
        "document": "00000000000",
        "bankAccount": {
            "bank": "001",
            "branch": "0001",
            "account": "12345",
        },
    }


def test_invalid_source_persists_nothing() -> None:
    _skip_if_no_dynamo_local()
    customer_id = generate_random_id()

    handle_request(
        {
            "source": "unrelated.source",
            "detail-type": PENDING_TRANSACTION_TYPE,
            "detail": _make_event_detail(customer_id, generate_random_id()),
        }
    )

    saved = get_deposit_by_user_id(customer_id)
    assert saved == []


def test_invalid_detail_type_persists_nothing() -> None:
    _skip_if_no_dynamo_local()
    customer_id = generate_random_id()

    handle_request(
        {
            "source": PENDING_TRANSACTION_SOURCE,
            "detail-type": "wrong-type",
            "detail": _make_event_detail(customer_id, generate_random_id()),
        }
    )

    saved = get_deposit_by_user_id(customer_id)
    assert saved == []


def test_invalid_json_persists_nothing() -> None:
    _skip_if_no_dynamo_local()
    customer_id = generate_random_id()

    handle_request(
        {
            "source": PENDING_TRANSACTION_SOURCE,
            "detail-type": PENDING_TRANSACTION_TYPE,
            "detail": "not-json",
        }
    )

    saved = get_deposit_by_user_id(customer_id)
    assert saved == []


def test_valid_event_persists_deposit() -> None:
    _skip_if_no_dynamo_local()
    customer_id = generate_random_id()
    transaction_id = generate_random_id()

    handle_request(
        {
            "source": PENDING_TRANSACTION_SOURCE,
            "detail-type": PENDING_TRANSACTION_TYPE,
            "detail": _make_event_detail(customer_id, transaction_id),
        }
    )

    saved = get_deposit_by_user_id(customer_id)
    assert len(saved) == 1
    d = saved[0]
    assert d.transaction.user_id == customer_id
    assert d.transaction.transaction_id == transaction_id
    assert d.transaction.amount == 100.0
    assert d.user_data is not None
    assert d.user_data.document == "00000000000"
    assert d.user_data.bank_account.bank == "001"
