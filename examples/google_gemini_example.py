"""
Google Gemini SDK proxied through probe.

Install:
    pip install google-genai

Start probe first:
    probe listen

Then run:
    GEMINI_API_KEY=your-key python examples/google_gemini_example.py
"""

import os
from google import genai
from google.genai import types

# Probe intercepts by pointing the SDK at localhost:9000
client = genai.Client(
    api_key=os.environ.get("GEMINI_API_KEY"),
    http_options=types.HttpOptions(base_url="http://localhost:9000"),
)

MODEL = "gemini-2.5-flash"  # Gemini 2.5 Flash — latest stable Gemini model


def non_streaming():
    print("=== Non-Streaming ===")
    response = client.models.generate_content(
        model=MODEL,
        contents="Say hello in 10 words.",
    )
    print(response.text)


def streaming():
    print("\n=== Streaming ===")
    for chunk in client.models.generate_content_stream(
        model=MODEL,
        contents="Count from 1 to 5 slowly.",
    ):
        print(chunk.text, end="", flush=True)
    print()


def with_system_prompt():
    print("\n=== System Prompt ===")
    response = client.models.generate_content(
        model=MODEL,
        config=types.GenerateContentConfig(
            system_instruction="You are a pirate. Always respond in pirate speak.",
        ),
        contents="What is the capital of France?",
    )
    print(response.text)


def with_tools():
    print("\n=== Tool Calls ===")
    def get_weather(city: str) -> str:
        """Get current weather for a city."""
        return f"Sunny, 22°C in {city}"

    response = client.models.generate_content(
        model=MODEL,
        contents="What's the weather in Tokyo?",
        config=types.GenerateContentConfig(tools=[get_weather]),
    )
    print(response.text)


if __name__ == "__main__":
    non_streaming()
    streaming()
    with_system_prompt()
    with_tools()
