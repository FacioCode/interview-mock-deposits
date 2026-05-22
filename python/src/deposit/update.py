from dataclasses import dataclass

from deposit.domain.deposit_status import DepositStatus


@dataclass
class InconsistentStatusChangeError(Exception):
    old_status: DepositStatus
    new_status: DepositStatus

    def __str__(self) -> str:
        return "inconsistent status change"


# TODO: implementar máquina de estados real (ex.: NEW -> DEPOSIT_SENT -> DONE/FAILED/RETURNED;
# proibir transições inválidas como FAILED -> DONE). Hoje aceita qualquer mudança != atual.
def validate_status_change(old_status: DepositStatus, new_status: DepositStatus) -> bool:
    return old_status != new_status


def update_deposit_status(deposit_id: str, new_status: DepositStatus) -> None:
    # TODO: replace with real implementation
    return None


def update_deposit_as_failed(deposit_id: str, reason: str) -> None:
    # TODO: replace with real implementation
    return None
