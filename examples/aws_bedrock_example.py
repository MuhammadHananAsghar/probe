"""
AWS Bedrock proxied through probe (via boto3 HTTP proxy).

Install:
    pip install boto3

Start probe first:
    probe listen

Then run:
    AWS_ACCESS_KEY_ID=your-key \
    AWS_SECRET_ACCESS_KEY=your-secret \
    AWS_DEFAULT_REGION=us-east-1 \
    python examples/aws_bedrock_example.py

Probe intercepts via HTTP_PROXY env var or botocore proxy config.
"""

import os
import json
import boto3
from botocore.config import Config

# Route boto3 traffic through probe via proxy config
proxy_config = Config(
    proxies={"https": "http://localhost:9000"},
)

client = boto3.client(
    "bedrock-runtime",
    region_name=os.environ.get("AWS_DEFAULT_REGION", "us-east-1"),
    config=proxy_config,
)

TITAN_MODEL_ID = "amazon.titan-text-lite-v1"
ANTHROPIC_MODEL_ID = "anthropic.claude-3-haiku-20240307-v1:0"


def invoke_titan():
    print("=== Amazon Titan (Non-Streaming) ===")
    body = json.dumps({
        "inputText": "Say hello in 10 words.",
        "textGenerationConfig": {"maxTokenCount": 64},
    })
    response = client.invoke_model(modelId=TITAN_MODEL_ID, body=body)
    result = json.loads(response["body"].read())
    print(result["results"][0]["outputText"])


def invoke_anthropic():
    print("\n=== Anthropic on Bedrock (Non-Streaming) ===")
    body = json.dumps({
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens": 64,
        "messages": [{"role": "user", "content": "Say hello in 10 words."}],
    })
    response = client.invoke_model(modelId=ANTHROPIC_MODEL_ID, body=body)
    result = json.loads(response["body"].read())
    print(result["content"][0]["text"])


def invoke_anthropic_streaming():
    print("\n=== Anthropic on Bedrock (Streaming) ===")
    body = json.dumps({
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens": 64,
        "messages": [{"role": "user", "content": "Count from 1 to 5 slowly."}],
    })
    response = client.invoke_model_with_response_stream(modelId=ANTHROPIC_MODEL_ID, body=body)
    for event in response["body"]:
        chunk = json.loads(event["chunk"]["bytes"])
        if chunk.get("type") == "content_block_delta":
            print(chunk["delta"].get("text", ""), end="", flush=True)
    print()


if __name__ == "__main__":
    invoke_titan()
    invoke_anthropic()
    invoke_anthropic_streaming()
