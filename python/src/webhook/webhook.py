from webhook.domain import DEPOSIT_OBJ, Body


def handle_webhook(body: Body) -> None:
    if body.object != DEPOSIT_OBJ:
        return None

    # TODO: complete the implementation

    return None
