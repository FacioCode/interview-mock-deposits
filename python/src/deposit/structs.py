import uuid
from dataclasses import dataclass, field

from deposit.domain.deposit_status import DepositStatus
from deposit.domain.transaction_type import TransactionType


@dataclass
class BankAccount:
    bank: str
    branch: str
    account: str

    def validate(self) -> None:
        if not self.bank:
            raise ValueError("invalid bank")
        if not self.branch:
            raise ValueError("invalid bank branch")
        if not self.account:
            raise ValueError("invalid bank account")


@dataclass
class UserData:
    name: str
    document: str
    bank_account: BankAccount

    def validate(self) -> None:
        if not self.document:
            raise ValueError("invalid document")
        if not self.name:
            raise ValueError("invalid name")
        self.bank_account.validate()


@dataclass
class Transaction:
    user_id: str
    type: TransactionType
    transaction_id: str
    amount: float
    delay_hours: int | None = None

    def generate_deposit_id(self) -> str:
        return str(uuid.uuid4())


@dataclass
class Deposit:
    deposit_id: str
    status: DepositStatus
    transaction: Transaction
    user_data: UserData | None = None
    created_at: str | None = field(default=None)
