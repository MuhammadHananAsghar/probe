# Probe Examples

Python SDK examples for every supported provider, all routed through probe.

## Setup

Start probe before running any example:

```bash
probe listen
```

## Examples

| File | Provider | SDK | Install |
|---|---|---|---|
| `openai_example.py` | OpenAI | `openai` | `pip install openai` |
| `anthropic_example.py` | Anthropic | `anthropic` | `pip install anthropic` |
| `google_gemini_example.py` | Google Gemini | `google-genai` | `pip install google-genai` |
| `groq_example.py` | Groq | `groq` | `pip install groq` |
| `mistral_example.py` | Mistral | `mistralai` | `pip install mistralai` |
| `cohere_example.py` | Cohere | `cohere` | `pip install cohere` |
| `together_example.py` | Together AI | `together` | `pip install together` |
| `fireworks_example.py` | Fireworks AI | `fireworks-ai` | `pip install fireworks-ai` |
| `openrouter_example.py` | OpenRouter | `openai` | `pip install openai` |
| `azure_openai_example.py` | Azure OpenAI | `openai` | `pip install openai` |
| `ollama_example.py` | Ollama (local) | `openai` | `pip install openai` |
| `aws_bedrock_example.py` | AWS Bedrock | `boto3` | `pip install boto3` |

## base_url reference

| Provider | base_url for probe |
|---|---|
| OpenAI | `http://localhost:9000/v1` |
| Anthropic | `http://localhost:9000` |
| Groq | `http://localhost:9000` |
| Mistral | `http://localhost:9000` (server_url) |
| Cohere | `http://localhost:9000` |
| Together AI | `http://localhost:9000/v1` |
| Fireworks AI | `http://localhost:9000/v1` |
| OpenRouter | `http://localhost:9000/v1` |
| Ollama | `http://localhost:9000/v1` |
| Google Gemini | `http://localhost:9000` (http_options) |
| Azure OpenAI | httpx proxy `http://localhost:9000` |
| AWS Bedrock | boto3 proxy `http://localhost:9000` |
