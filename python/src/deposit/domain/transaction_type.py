from enum import Enum
from typing import Any


class TransactionType(Enum):
    SALARY_ADVANCE = 0


_TYPE_TO_STRING: dict[TransactionType, str] = {
    TransactionType.SALARY_ADVANCE: "SalaryAdvance",
}

_STRING_TO_TYPE: dict[str, TransactionType] = {
    "SalaryAdvance": TransactionType.SALARY_ADVANCE,
}


def transaction_type_to_string(t: TransactionType) -> str:
    return _TYPE_TO_STRING[t]


def transaction_type_from_string(value: str) -> TransactionType:
    try:
        return _STRING_TO_TYPE[value]
    except KeyError as exc:
        raise ValueError(f"invalid TransactionType: {value}") from exc


def transaction_type_to_ddb(t: TransactionType) -> dict[str, Any]:
    return {"S": _TYPE_TO_STRING[t]}


def transaction_type_from_ddb(attr: dict[str, Any]) -> TransactionType:
    value = attr.get("S")
    if value is None:
        raise ValueError("missing TransactionType attribute")
    return transaction_type_from_string(value)
