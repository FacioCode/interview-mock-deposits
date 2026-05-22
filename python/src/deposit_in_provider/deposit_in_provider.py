import logging

from deposit import get_deposit_by_id
from deposit.domain.deposit_status import DepositStatus
from deposit_in_provider.pay_user import pay_user
from internal.events import DepositRequestedEvent

log = logging.getLogger(__name__)


def create_deposit_in_provider(event: DepositRequestedEvent, request_id: str) -> None:
    deposit_id = event.id
    try:
        dep = get_deposit_by_id(deposit_id)
    except Exception as exc:
        log.warning("error getting deposit", extra={"error": str(exc)})
        raise

    if dep.status != DepositStatus.NEW:
        log.warning(
            "deposit not new",
            extra={"depositId": deposit_id, "status": dep.status.name},
        )
        return None

    if dep.user_data is None:
        log.error("deposit user data is nil", extra={"depositId": deposit_id})
        return None

    return pay_user(dep, request_id)
