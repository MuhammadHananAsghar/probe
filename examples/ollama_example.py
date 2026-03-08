"""
Ollama proxied through probe (uses OpenAI-compatible SDK).

Install:
    pip install openai
    # Also install Ollama: https://ollama.com
    # Pull a model: ollama pull llama3.2

Start probe first (probe will forward to your local Ollama):
    probe listen

Then run:
    python examples/ollama_example.py

Note: Ollama runs locally so no API key is needed.
Probe detects Ollama via the /api/ path prefix.
"""

from openai import OpenAI

# Ollama's native endpoint — probe intercepts and detects it as Ollama
client = OpenAI(
    api_key="ollama",  # Ollama ignores the key but SDK requires a non-empty value
    base_url="http://localhost:9000/v1",  # probe proxy → forwards to ollama at :11434
)

MODEL = "llama3.2"  # must be pulled: ollama pull llama3.2


def list_models():
    print("=== Available Ollama Models ===")
    models = client.models.list()
    for m in models.data:
        print(f"  {m.id}")


def non_streaming():
    print("\n=== Non-Streaming ===")
    response = client.chat.completions.create(
        model=MODEL,
        messages=[{"role": "user", "content": "Say hello in 10 words."}],
        max_tokens=64,
    )
    print(response.choices[0].message.content)


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
    list_models()
    non_streaming()
    streaming()
    with_system_prompt()
