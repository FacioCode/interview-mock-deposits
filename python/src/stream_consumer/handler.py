import logging
from typing import Any

from stream_consumer.consumer import consume

log = logging.getLogger()
log.setLevel(logging.INFO)


def handle_request(event: dict[str, Any], context: Any | None = None) -> None:
    """Entry point for the stream_consumer Lambda.

    Receives a DynamoDB stream batch from Lambda and dispatches each record.
    Event shape: {"Records": [{ "eventName": "INSERT"|"MODIFY"|..., "dynamodb": {...}}, ...]}
    """
    records = event.get("Records", [])
    log.info("stream consumer starts", extra={"recordCount": len(records)})
    consume(records)
