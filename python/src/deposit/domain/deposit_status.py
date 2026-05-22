from enum import Enum
from typing import Any


class DepositStatus(Enum):
    NEW = 0
    DEPOSIT_SENT = 1
    FAILED = 2


_STATUS_TO_STRING: dict[DepositStatus, str] = {
    DepositStatus.NEW: "NEW",
    DepositStatus.DEPOSIT_SENT: "DEPOSIT_SENT",
    DepositStatus.FAILED: "FAILED",
}

_STRING_TO_STATUS: dict[str, DepositStatus] = {
    "NEW": DepositStatus.NEW,
    "DEPOSIT_SENT": DepositStatus.DEPOSIT_SENT,
    "FAILED": DepositStatus.FAILED,
}


def deposit_status_to_string(status: DepositStatus) -> str:
    return _STATUS_TO_STRING[status]


def deposit_status_from_string(value: str) -> DepositStatus:
    try:
        return _STRING_TO_STATUS[value]
    except KeyError as exc:
        raise ValueError(f"invalid DepositStatus: {value}") from exc


def deposit_status_to_ddb(status: DepositStatus) -> dict[str, Any]:
    return {"S": _STATUS_TO_STRING[status]}


def deposit_status_from_ddb(attr: dict[str, Any]) -> DepositStatus:
    value = attr.get("S")
    if value is None:
        raise ValueError("missing DepositStatus attribute")
    return deposit_status_from_string(value)
