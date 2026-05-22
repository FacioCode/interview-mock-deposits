import os
from typing import Any

import boto3


def new_dynamo_client() -> Any:
    """Build a DynamoDB client.

    If AWS_ENDPOINT_URL_DYNAMODB is set (tests via with-dynamodb-local.sh),
    point there with dummy credentials. Otherwise use the standard SDK
    credential chain.
    """
    endpoint = os.getenv("AWS_ENDPOINT_URL_DYNAMODB")
    if endpoint:
        return boto3.client(
            "dynamodb",
            endpoint_url=endpoint,
            region_name="local",
            aws_access_key_id="dummy",
            aws_secret_access_key="dummy",
        )
    return boto3.client("dynamodb")
