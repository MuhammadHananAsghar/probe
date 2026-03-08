"""
Together AI SDK proxied through probe.

Install:
    pip install together

Start probe first:
    probe listen

Then run:
    TOGETHER_API_KEY=your-key python examples/together_example.py
"""

import os
from together import Together

client = Together(
    api_key=os.environ.get("TOGETHER_API_KEY"),
    base_url="http://localhost:9000/v1",  # probe proxy
)

MODEL = "meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8"  # Llama 4 Maverick — latest on Together AI


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


def with_system_prompt():
    print("\n=== System Prompt ===")
    response = client.chat.completions.create(
        model=MODEL,
        messages=[
            {"role": "system", "content": "You are a pirate. Always respond in pirate speak."},
            {"role": "user", "content": "What is the capital of France?"},
        ],
        max_tokens=128,
    )
    print(response.choices[0].message.content)


if __name__ == "__main__":
    non_streaming()
    streaming()
    with_system_prompt()
