import uuid


def generate_random_id() -> str:
    return str(uuid.uuid4())
