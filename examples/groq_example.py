"""
Groq SDK proxied through probe.

Install:
    pip install groq

Start probe first:
    probe listen

Then run:
    GROQ_API_KEY=your-key python examples/groq_example.py
"""

import os
from groq import Groq

client = Groq(
    api_key=os.environ.get("GROQ_API_KEY"),
    base_url="http://localhost:9000",  # probe proxy
)

MODEL = "meta-llama/llama-4-maverick-17b-128e-instruct"  # Llama 4 Maverick — latest on Groq


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


def with_tools():
    print("\n=== Tool Calls ===")
    tools = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get current weather for a city.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "city": {"type": "string", "description": "City name"},
                    },
                    "required": ["city"],
                },
            },
        }
    ]
    response = client.chat.completions.create(
        model=MODEL,
        messages=[{"role": "user", "content": "What's the weather in Tokyo?"}],
        tools=tools,
        tool_choice="auto",
    )
    msg = response.choices[0].message
    if msg.tool_calls:
        for tc in msg.tool_calls:
            print(f"Tool: {tc.function.name}  Args: {tc.function.arguments}")
    else:
        print(msg.content)


if __name__ == "__main__":
    non_streaming()
    streaming()
    with_tools()
