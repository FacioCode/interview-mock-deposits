from dataclasses import dataclass
from typing import Any

DONE = "DONE"
FAILED = "FAILED"
RETURNED = "RETURNED"
DEPOSIT_OBJ = "Deposits"


@dataclass
class Data:
    id: str
    status: str
    status_description: str | None
    integration_id: str
    bank_receipt_url: str
    authorization_code: str | None = None

    @classmethod
    def from_json(cls, raw: dict[str, Any]) -> "Data":
        return cls(
            id=raw.get("id", ""),
            status=raw.get("status", ""),
            status_description=raw.get("status_description"),
            integration_id=raw.get("integration_id", ""),
            bank_receipt_url=raw.get("bank_receipt_url", ""),
            authorization_code=raw.get("authorization_code"),
        )


@dataclass
class Body:
    id: str
    object: str
    date: str
    data: Data
    version: str

    @classmethod
    def from_json(cls, raw: dict[str, Any]) -> "Body":
        return cls(
            id=raw.get("id", ""),
            object=raw.get("object", ""),
            date=raw.get("date", ""),
            data=Data.from_json(raw.get("data", {})),
            version=raw.get("version", ""),
        )
