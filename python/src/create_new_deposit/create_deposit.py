import logging

from deposit import (
    BankAccount,
    Transaction,
    UserData,
    create_deposit_with_user_data,
)
from deposit.domain.transaction_type import TransactionType
from internal.events import PendingTransactionEvent

log = logging.getLogger(__name__)


def create_deposit(event: PendingTransactionEvent, request_id: str) -> None:
    log.info(
        "creating deposit for pending transaction",
        extra={"transactionId": event.transaction_id, "customerId": event.customer_id},
    )

    amount = float(event.amount)

    transaction = Transaction(
        user_id=event.customer_id,
        type=TransactionType.SALARY_ADVANCE,
        transaction_id=event.transaction_id,
        amount=amount,
    )

    user_data = UserData(
        name=event.name,
        document=event.document,
        bank_account=BankAccount(
            bank=event.bank_account.bank,
            branch=event.bank_account.branch,
            account=event.bank_account.account,
        ),
    )

    create_deposit_with_user_data(transaction, user_data, request_id)
