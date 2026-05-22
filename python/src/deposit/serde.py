from decimal import Decimal
from typing import Any

from deposit.domain.deposit_status import deposit_status_from_ddb, deposit_status_to_ddb
from deposit.domain.transaction_type import transaction_type_from_ddb, transaction_type_to_ddb
from deposit.structs import BankAccount, Deposit, Transaction, UserData


def _bank_account_to_ddb(ba: BankAccount) -> dict[str, Any]:
    return {
        "M": {
            "bank": {"S": ba.bank},
            "branch": {"S": ba.branch},
            "account": {"S": ba.account},
        }
    }


def _bank_account_from_ddb(attr: dict[str, Any]) -> BankAccount:
    m = attr["M"]
    return BankAccount(
        bank=m["bank"]["S"],
        branch=m["branch"]["S"],
        account=m["account"]["S"],
    )


def _user_data_to_ddb(ud: UserData) -> dict[str, Any]:
    return {
        "M": {
            "name": {"S": ud.name},
            "document": {"S": ud.document},
            "bankAccount": _bank_account_to_ddb(ud.bank_account),
        }
    }


def _user_data_from_ddb(attr: dict[str, Any]) -> UserData:
    m = attr["M"]
    return UserData(
        name=m["name"]["S"],
        document=m["document"]["S"],
        bank_account=_bank_account_from_ddb(m["bankAccount"]),
    )


def deposit_to_ddb_item(deposit: Deposit) -> dict[str, Any]:
    item: dict[str, Any] = {
        "depositId": {"S": deposit.deposit_id},
        "status": deposit_status_to_ddb(deposit.status),
        "userId": {"S": deposit.transaction.user_id},
        "type": transaction_type_to_ddb(deposit.transaction.type),
        "transactionId": {"S": deposit.transaction.transaction_id},
        "amount": {"N": str(Decimal(str(deposit.transaction.amount)))},
    }
    if deposit.transaction.delay_hours is not None:
        item["delayHours"] = {"N": str(deposit.transaction.delay_hours)}
    if deposit.user_data is not None and deposit.user_data.document:
        item["userData"] = _user_data_to_ddb(deposit.user_data)
    if deposit.created_at:
        item["createdAt"] = {"S": deposit.created_at}
    return item


def deposit_from_ddb_item(item: dict[str, Any]) -> Deposit:
    transaction = Transaction(
        user_id=item["userId"]["S"],
        type=transaction_type_from_ddb(item["type"]),
        transaction_id=item["transactionId"]["S"],
        amount=float(item["amount"]["N"]),
        delay_hours=int(item["delayHours"]["N"]) if "delayHours" in item else None,
    )
    user_data: UserData | None = None
    if "userData" in item:
        user_data = _user_data_from_ddb(item["userData"])
    return Deposit(
        deposit_id=item["depositId"]["S"],
        status=deposit_status_from_ddb(item["status"]),
        transaction=transaction,
        user_data=user_data,
        created_at=item.get("createdAt", {}).get("S"),
    )
