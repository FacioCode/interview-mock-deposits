from dataclasses import dataclass
from typing import Any

# Deposit lifecycle events (published by stream_consumer on DDB transitions)
SOURCE = "interview.mock.deposits"
DEPOSIT_REQUESTED_TYPE = "deposit-requested"

# Upstream events that trigger a new deposit
PENDING_TRANSACTION_SOURCE = "interview.mock.transactions"
PENDING_TRANSACTION_TYPE = "transaction-pending"


@dataclass
class BankAccount:
    bank: str
    branch: str
    account: str

    @classmethod
    def from_json(cls, raw: dict[str, Any]) -> "BankAccount":
        return cls(bank=raw["bank"], branch=raw["branch"], account=raw["account"])


@dataclass
class PendingTransactionEvent:
    transaction_id: str
    customer_id: str
    amount: str
    name: str
    document: str
    bank_account: BankAccount

    @classmethod
    def from_json(cls, raw: dict[str, Any]) -> "PendingTransactionEvent":
        return cls(
            transaction_id=raw["transactionId"],
            customer_id=raw["customerId"],
            amount=raw["amount"],
            name=raw["name"],
            document=raw["document"],
            bank_account=BankAccount.from_json(raw["bankAccount"]),
        )


@dataclass
class DepositRequestedEvent:
    id: str
    transaction_id: str
    customer_id: str

    @classmethod
    def from_json(cls, raw: dict[str, Any]) -> "DepositRequestedEvent":
        return cls(
            id=raw["id"],
            transaction_id=raw["transactionId"],
            customer_id=raw["customerId"],
        )
