"""
OpenRouter proxied through probe (uses OpenAI-compatible SDK).

Install:
    pip install openai

Start probe first:
    probe listen

Then run:
    OPENROUTER_API_KEY=your-key python examples/openrouter_example.py

OpenRouter is OpenAI-compatible, so we use the OpenAI SDK with a custom base_url.
"""

import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ.get("OPENROUTER_API_KEY"),
    base_url="http://localhost:9000/v1",  # probe proxy
    default_headers={
        "HTTP-Referer": "https://github.com/MuhammadHananAsghar/probe",
        "X-Title": "probe",
    },
)

MODEL = "openai/gpt-4.1-mini"  # OpenRouter model format: provider/model


def non_streaming():
    print("=== Non-Streaming ===")
    response = client.chat.completions.create(
        model=MODEL,
        messages=[{"role": "user", "content": "Say hello in 10 words."}],
        max_tokens=64,
    )
    print(response.choices[0].message.content)
    print(f"Tokens: {response.usage.total_tokens}")


def streaming():
    print("\n=== Streaming ===")
    stream = client.chat.completions.create(
        model=MODEL,
        messages=[{"role": "user", "content": "Count from 1 to 5 slowly."}],
        max_tokens=64,
        stream=True,
    )
    for chunk in stream:
        delta = chunk.choices[0].delta.content
        if delta:
            print(delta, end="", flush=True)
    print()


def multi_provider_compare():
    """Run the same prompt through different providers via OpenRouter."""
    print("\n=== Multi-Provider via OpenRouter ===")
    models = [
        "openai/gpt-4.1-mini",
        "anthropic/claude-haiku-4-5",
        "google/gemini-2.5-flash",
        "meta-llama/llama-4-maverick",
    ]
    prompt = "What is 2+2? Answer in one word."
    for model in models:
        resp = client.chat.completions.create(
            model=model,
            messages=[{"role": "user", "content": prompt}],
            max_tokens=16,
        )
        print(f"  {model:<45} → {resp.choices[0].message.content.strip()}")


if __name__ == "__main__":
    non_streaming()
    streaming()
    multi_provider_compare()
