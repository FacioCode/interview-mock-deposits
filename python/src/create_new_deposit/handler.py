import json
import logging
from typing import Any

from create_new_deposit.create_deposit import create_deposit
from internal.events import (
    PENDING_TRANSACTION_SOURCE,
    PENDING_TRANSACTION_TYPE,
    PendingTransactionEvent,
)

log = logging.getLogger()
log.setLevel(logging.INFO)


def handle_request(event: dict[str, Any], context: Any | None = None) -> None:
    """Entry point for the create_new_deposit Lambda.

    Accepts an EventBridge / CloudWatch event payload, validates the source
    and detail-type, then dispatches to create_deposit. Returns None to
    match the Go reference's "swallow & log errors" behavior.
    """
    request_id = getattr(context, "aws_request_id", "") if context is not None else ""

    log.info("event received", extra={"event": event})

    try:
        source = event.get("source")
        if source != PENDING_TRANSACTION_SOURCE:
            log.error("invalid source: %s", source)
            return None

        detail_type = event.get("detail-type") or event.get("detailType")
        if detail_type != PENDING_TRANSACTION_TYPE:
            log.error("invalid event type: %s", detail_type)
            return None

        detail = event.get("detail")
        try:
            parsed = detail if isinstance(detail, dict) else json.loads(detail)
            pending = PendingTransactionEvent.from_json(parsed)
        except (json.JSONDecodeError, KeyError, TypeError) as exc:
            log.error("invalid pending transaction detail", extra={"error": str(exc)})
            return None

        create_deposit(pending, request_id)
        return None
    except Exception:
        log.exception("panic in create_new_deposit handler")
        raise
