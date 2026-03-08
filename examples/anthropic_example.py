"""
Anthropic SDK proxied through probe.

Start probe first:
    probe listen

Then run:
    ANTHROPIC_API_KEY=your-key python examples/anthropic_example.py
"""

import os
import anthropic

client = anthropic.Anthropic(
    api_key=os.environ.get("ANTHROPIC_API_KEY"),
    base_url="http://localhost:9000",  # probe proxy (no /v1 suffix for Anthropic)
)


def non_streaming():
    print("=== Non-Streaming ===")
    response = client.messages.create(
        model="claude-haiku-4-5-20251001",
        max_tokens=64,
        messages=[{"role": "user", "content": "Say hello in 10 words."}],
    )
    print(response.content[0].text)
    print(f"Tokens: input={response.usage.input_tokens} output={response.usage.output_tokens}")


def streaming():
    print("\n=== Streaming ===")
    with client.messages.stream(
        model="claude-haiku-4-5-20251001",
        max_tokens=64,
        messages=[{"role": "user", "content": "Count from 1 to 5 slowly."}],
    ) as stream:
        for text in stream.text_stream:
            print(text, end="", flush=True)
    print()


def with_tools():
    print("\n=== Tool Calls ===")
    tools = [
        {
            "name": "get_weather",
            "description": "Get current weather for a city.",
            "input_schema": {
                "type": "object",
                "properties": {
                    "city": {"type": "string", "description": "City name"},
                },
                "required": ["city"],
            },
        }
    ]
    response = client.messages.create(
        model="claude-haiku-4-5-20251001",
        max_tokens=256,
        tools=tools,
        messages=[{"role": "user", "content": "What's the weather in Tokyo?"}],
    )
    for block in response.content:
        if block.type == "tool_use":
            print(f"Tool: {block.name}  Input: {block.input}")
        elif block.type == "text":
            print(block.text)


def with_system_prompt():
    print("\n=== System Prompt ===")
    response = client.messages.create(
        model="claude-haiku-4-5-20251001",
        max_tokens=128,
        system="You are a pirate. Always respond in pirate speak.",
        messages=[{"role": "user", "content": "What is the capital of France?"}],
    )
    print(response.content[0].text)


if __name__ == "__main__":
    non_streaming()
    streaming()
    with_tools()
    with_system_prompt()
