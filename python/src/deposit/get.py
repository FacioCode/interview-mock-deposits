import os
from typing import Any

from deposit.serde import deposit_from_ddb_item
from deposit.structs import Deposit
from utils.awsutil import new_dynamo_client


def _table_name() -> str:
    name = os.environ.get("TABLE_NAME")
    if not name:
        raise RuntimeError("TABLE_NAME env var is required")
    return name


def _ddb() -> Any:
    return new_dynamo_client()


def get_deposit_by_user_id(user_id: str) -> list[Deposit]:
    response = _ddb().query(
        TableName=_table_name(),
        IndexName="userIndex",
        KeyConditions={
            "userId": {
                "ComparisonOperator": "EQ",
                "AttributeValueList": [{"S": user_id}],
            }
        },
    )
    return [deposit_from_ddb_item(item) for item in response.get("Items", [])]


def get_deposit_by_id(deposit_id: str) -> Deposit:
    if not deposit_id:
        raise ValueError("Deposit not found")
    response = _ddb().get_item(
        TableName=_table_name(),
        Key={"depositId": {"S": deposit_id}},
        ConsistentRead=True,
    )
    item = response.get("Item")
    if not item:
        raise ValueError("Deposit not found")
    return deposit_from_ddb_item(item)
