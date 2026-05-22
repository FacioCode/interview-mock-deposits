import json
import logging
from typing import Any

from deposit_in_provider.deposit_in_provider import create_deposit_in_provider
from internal.events import (
    DEPOSIT_REQUESTED_TYPE,
    SOURCE,
    DepositRequestedEvent,
)

log = logging.getLogger()
log.setLevel(logging.INFO)


def handle_request(event: dict[str, Any], context: Any | None = None) -> None:
    """Entry point for the deposit_in_provider Lambda.

    aws_request_id is pulled off the context and forwarded into the domain
    flow — what pay_user does with it is a design decision.
    """
    aws_request_id = getattr(context, "aws_request_id", "") if context is not None else ""

    log.info("event received", extra={"event": event})

    source = event.get("source")
    if source != SOURCE:
        log.error("invalid source: %s", source)
        return None

    detail_type = event.get("detail-type") or event.get("detailType")
    if detail_type != DEPOSIT_REQUESTED_TYPE:
        log.error("invalid event type: %s", detail_type)
        return None

    detail = event.get("detail")
    try:
        parsed = detail if isinstance(detail, dict) else json.loads(detail)
        deposit_requested = DepositRequestedEvent.from_json(parsed)
    except (json.JSONDecodeError, KeyError, TypeError) as exc:
        log.error("invalid deposit requested detail", extra={"error": str(exc)})
        return None

    try:
        create_deposit_in_provider(deposit_requested, aws_request_id)
    except Exception as exc:
        log.error(
            "error creating deposit in provider",
            extra={"event": event, "error": str(exc)},
        )
        return None
    return None
