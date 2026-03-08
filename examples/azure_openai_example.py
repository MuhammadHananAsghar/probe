"""
Azure OpenAI SDK proxied through probe.

Install:
    pip install openai

Start probe first:
    probe listen

Then run:
    AZURE_OPENAI_API_KEY=your-key \
    AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com \
    AZURE_OPENAI_DEPLOYMENT=your-deployment-name \
    python examples/azure_openai_example.py
"""

import os
from openai import AzureOpenAI

ENDPOINT = os.environ.get("AZURE_OPENAI_ENDPOINT", "https://your-resource.openai.azure.com")
DEPLOYMENT = os.environ.get("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-mini")

# AzureOpenAI doesn't support base_url directly, so we route via probe by
# replacing the endpoint host with probe's address.
# Probe detects Azure via the "openai.azure.com" hostname in the Host header.
client = AzureOpenAI(
    api_key=os.environ.get("AZURE_OPENAI_API_KEY"),
    azure_endpoint=ENDPOINT,
    api_version="2024-10-21",
    http_client=__import__("httpx").Client(
        transport=__import__("httpx").HTTPTransport(proxy="http://localhost:9000"),
    ),
)


def non_streaming():
    print("=== Non-Streaming ===")
    response = client.chat.completions.create(
        model=DEPLOYMENT,
        messages=[{"role": "user", "content": "Say hello in 10 words."}],
        max_tokens=64,
    )
    print(response.choices[0].message.content)
    print(f"Tokens: {response.usage.total_tokens}")


def streaming():
    print("\n=== Streaming ===")
    stream = client.chat.completions.create(
        model=DEPLOYMENT,
        messages=[{"role": "user", "content": "Count from 1 to 5 slowly."}],
        max_tokens=64,
        stream=True,
    )
    for chunk in stream:
        if chunk.choices:
            delta = chunk.choices[0].delta.content
            if delta:
                print(delta, end="", flush=True)
    print()


if __name__ == "__main__":
    non_streaming()
    streaming()
