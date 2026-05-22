import hmac
import json
import os
from dataclasses import dataclass
from hashlib import sha256
from typing import Any, Literal
from urllib import error, request

DUPLICATED_TRANSFER_ERROR_CODE = "PARTNER_DUPLICATE"
BLOCKED_DOCUMENT_ERROR_CODE = "PARTNER_BLOCKED_DOCUMENT"

STATUS_PENDING = "PENDING"
STATUS_CONFIRMED = "CONFIRMED"


class PartnerUnavailableError(Exception):
    """The partner endpoint could not be reached."""


class InvalidCredentialsError(Exception):
    """Partner credentials are missing or invalid."""


class BlockedDocumentError(Exception):
    """The destination document is blocked by the partner."""


class TransferAlreadyClosedError(Exception):
    """The partner reports the transfer was already closed (idempotent)."""


@dataclass
class PartnerSecrets:
    api_key: str
    api_secret: str


@dataclass
class TransferRequest:
    id: str
    amount: float


@dataclass
class TransferResult:
    id: str
    status: Literal["CONFIRMED", "PENDING"]


def _partner_base_url() -> str:
    return os.getenv("PARTNER_BASE_URL", "https://api.partner.example.com")


def get_partner_secrets() -> PartnerSecrets:
    """Load partner credentials from env. In production these would come from
    Secrets Manager / KMS."""
    api_key = os.getenv("PARTNER_API_KEY", "")
    api_secret = os.getenv("PARTNER_API_SECRET", "")
    if not api_key or not api_secret:
        raise InvalidCredentialsError("invalid partner credentials")
    return PartnerSecrets(api_key=api_key, api_secret=api_secret)


def _http_request(method: str, url: str, *, headers: dict[str, str], body: bytes | None) -> tuple[int, bytes]:
    req = request.Request(url, data=body, method=method, headers=headers)
    try:
        with request.urlopen(req, timeout=30) as resp:
            return resp.status, resp.read()
    except error.HTTPError as exc:
        return exc.code, exc.read()
    except (error.URLError, TimeoutError, OSError) as exc:
        raise PartnerUnavailableError("partner API unavailable") from exc


def _create_transfer(secrets: PartnerSecrets, req: TransferRequest) -> str:
    body = json.dumps({"id": req.id, "amount": req.amount}).encode("utf-8")
    headers = {"Content-Type": "application/json", "X-Api-Key": secrets.api_key}
    status, raw = _http_request("POST", _partner_base_url() + "/transfers", headers=headers, body=body)
    if 200 <= status < 300:
        out = json.loads(raw or b"{}")
        return str(out.get("transferId", ""))

    err_body: dict[str, Any] = {}
    if raw:
        try:
            err_body = json.loads(raw)
        except json.JSONDecodeError:
            pass
    code = err_body.get("code")
    message = err_body.get("message", "")
    if code == DUPLICATED_TRANSFER_ERROR_CODE:
        raise RuntimeError(f"duplicated transfer: {message}")
    if code == BLOCKED_DOCUMENT_ERROR_CODE:
        raise BlockedDocumentError("destination document is blocked")
    raise RuntimeError(f"create transfer status {status}: {message}")


def _close_transfer(secrets: PartnerSecrets, transfer_id: str) -> None:
    headers = {"X-Api-Key": secrets.api_key}
    url = f"{_partner_base_url()}/transfers/{transfer_id}/close"
    status, _ = _http_request("POST", url, headers=headers, body=None)
    if status == 409:
        raise TransferAlreadyClosedError("transfer already closed")
    if 200 <= status < 300:
        return
    raise RuntimeError(f"close transfer status {status}")


class PendingCloseError(Exception):
    """Raised when the partner transfer was created but close failed.

    The transfer exists on the partner side; the caller should persist
    DEPOSIT_SENT (or equivalent) and wait for the webhook, rather than
    marking the deposit as failed.
    """

    def __init__(self, result: TransferResult, cause: Exception) -> None:
        super().__init__(f"close transfer: {cause}")
        self.result = result
        self.cause = cause


def pay(req: TransferRequest) -> TransferResult:
    """Run the two-step partner flow: create a pending transfer, then close it.

    Returns CONFIRMED on full success. If create succeeded but close failed,
    returns the transferId with status PENDING so the caller can decide whether
    to retry only the close step.
    """
    secrets = get_partner_secrets()

    transfer_id = _create_transfer(secrets, req)

    try:
        _close_transfer(secrets, transfer_id)
    except Exception as exc:
        # Create succeeded but close failed. The transfer exists on the
        # partner side, so surface PENDING + the underlying error.
        result = TransferResult(id=transfer_id, status=STATUS_PENDING)
        raise PendingCloseError(result, exc) from exc

    return TransferResult(id=transfer_id, status=STATUS_CONFIRMED)


def is_webhook_valid(signature: str, body: str) -> bool:
    """Verify an HMAC-SHA256 signature against the raw webhook body.

    The partner signs the body with api_secret and sends the hex digest in
    the x-partner-signature header.
    """
    if not signature:
        return False
    try:
        secrets = get_partner_secrets()
    except InvalidCredentialsError:
        return False
    mac = hmac.new(secrets.api_secret.encode("utf-8"), body.encode("utf-8"), sha256)
    return hmac.compare_digest(mac.hexdigest(), signature)
