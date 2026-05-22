import logging
from typing import Any

log = logging.getLogger(__name__)


def send_event(detail_type: str, item: dict[str, Any]) -> None:
    # mock event publication (EventBridge PutEvents)
    deposit_id = item.get("depositId", {}).get("S", "")
    log.info(
        "event would be published",
        extra={"detailType": detail_type, "depositId": deposit_id},
    )
