import os
from datetime import datetime, timezone
from typing import Any

from botocore.exceptions import ClientError

from deposit.domain.deposit_status import DepositStatus
from deposit.serde import deposit_to_ddb_item
from deposit.structs import Deposit, Transaction, UserData
from utils.awsutil import new_dynamo_client


class ErrDepositAlreadyExists(Exception):
    """Raised when a PutItem fails the depositId uniqueness condition."""


def _table_name() -> str:
    name = os.environ.get("TABLE_NAME")
    if not name:
        raise RuntimeError("TABLE_NAME env var is required")
    return name


def _ddb() -> Any:
    return new_dynamo_client()


def create_deposit(request: Transaction, request_id: str) -> None:
    _create(request, None, request_id, DepositStatus.NEW)


def create_deposit_with_user_data(request: Transaction, user_data: UserData, request_id: str) -> None:
    _create(request, user_data, request_id, DepositStatus.NEW)


def create_deposit_with_status(request: Transaction, request_id: str, status: DepositStatus) -> None:
    _create(request, None, request_id, status)


def _create(
    request: Transaction,
    user_data: UserData | None,
    request_id: str,
    status: DepositStatus,
) -> None:
    new = Deposit(
        deposit_id=request.generate_deposit_id(),
        status=status,
        transaction=request,
    )
    if user_data is not None and user_data.document:
        new.user_data = user_data
    save_deposit(new, request_id)


def save_deposit(data: Deposit, request_id: str) -> None:
    item = deposit_to_ddb_item(data)
    item["createdAt"] = {"S": datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")}
    try:
        _ddb().put_item(
            TableName=_table_name(),
            Item=item,
            ConditionExpression="attribute_not_exists(depositId)",
        )
    except ClientError as exc:
        if exc.response.get("Error", {}).get("Code") == "ConditionalCheckFailedException":
            raise ErrDepositAlreadyExists("deposit already exists") from exc
        raise
