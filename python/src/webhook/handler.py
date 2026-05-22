import json
import logging
from typing import Any

from deposit import InconsistentStatusChangeError
from partner_api import is_webhook_valid
from webhook.domain import Body
from webhook.webhook import handle_webhook

log = logging.getLogger()
log.setLevel(logging.INFO)


def _lookup_header(headers: dict[str, str] | None, name: str) -> str:
    if not headers:
        return ""
    lower = name.lower()
    for key, value in headers.items():
        if key.lower() == lower:
            return value
    return ""


def handle_request(event: dict[str, Any], context: Any | None = None) -> dict[str, Any]:
    """Entry point for the webhook Lambda (API Gateway proxy integration).

    Signature validation via partner_api.is_webhook_valid is wired before
    any TODO so candidates consume rather than reimplement.
    """
    log.info("webhook request received", extra={"request": event})

    headers = event.get("headers") or {}
    raw_body = event.get("body") or ""
    signature = _lookup_header(headers, "x-partner-signature")

    if not signature or not is_webhook_valid(signature, raw_body):
        log.warning("invalid webhook request")
        return {"statusCode": 403, "body": ""}

    try:
        parsed = json.loads(raw_body) if raw_body else {}
        body = Body.from_json(parsed)
    except (json.JSONDecodeError, KeyError, TypeError) as exc:
        log.error("invalid webhook body", extra={"error": str(exc)})
        return {"statusCode": 500, "body": ""}

    try:
        handle_webhook(body)
    except InconsistentStatusChangeError as exc:
        log.critical(
            "inconsistent status change",
            extra={
                "depositId": body.data.integration_id,
                "depositProviderStatus": body.data.status,
                "error": str(exc),
            },
        )
        return {"statusCode": 200, "body": ""}
    except Exception as exc:
        log.error(
            "webhook error",
            extra={
                "depositId": body.data.integration_id,
                "depositProviderStatus": body.data.status,
                "error": str(exc),
            },
        )
        return {"statusCode": 500, "body": ""}

    return {"statusCode": 200, "body": ""}
