import logging

from deposit import Deposit

log = logging.getLogger(__name__)


def pay_user(deposit: Deposit, request_id: str) -> None:
    log.info("paying user via partner_api", extra={"depositId": deposit.deposit_id})
    # TODO: implement real payment logic
    return None
