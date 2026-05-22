from typing import Any

from deposit.domain.deposit_status import deposit_status_to_string
from deposit.domain.deposit_status import DepositStatus

INSERT_EVENT = "INSERT"
MODIFY_EVENT = "MODIFY"


def consume(records: list[dict[str, Any]]) -> None:
    for record in records:
        send_events(record)


def send_events(record: dict[str, Any]) -> None:
    from stream_consumer.events import send_event

    if is_new_deposit(record):
        send_event("deposit-requested", record["dynamodb"]["NewImage"])
    elif is_deposit_sent(record):
        send_event("deposit-sent", record["dynamodb"]["NewImage"])


def is_new_deposit(record: dict[str, Any]) -> bool:
    new_item = record.get("dynamodb", {}).get("NewImage", {})
    status_attr = new_item.get("status")
    if status_attr is None:
        return False
    return (
        record.get("eventName") == INSERT_EVENT
        and status_attr.get("S") == deposit_status_to_string(DepositStatus.NEW)
    )


def is_deposit_sent(record: dict[str, Any]) -> bool:
    new_item = record.get("dynamodb", {}).get("NewImage", {})
    old_item = record.get("dynamodb", {}).get("OldImage", {})
    new_status = new_item.get("status")
    if new_status is None:
        return False
    return (
        record.get("eventName") == MODIFY_EVENT
        and old_item.get("status", {}).get("S") != new_status.get("S")
        and new_status.get("S") == deposit_status_to_string(DepositStatus.DEPOSIT_SENT)
    )
