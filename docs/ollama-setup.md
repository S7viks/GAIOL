# Ollama Local Fallback Setup

## What is Ollama?

Ollama allows you to run large language models locally on your machine. This means:
- **No API rate limits!** 
- **No API costs!**
- **Works offline**
- **100% private - data never leaves your machine**

## Quick Setup (5 minutes)

### 1. Install Ollama

**Windows/Mac/Linux:**
Visit https://ollama.com and download the installer for your platform.

### 2. Download a Model

After installation, open a terminal/PowerShell and run:

```powershell
# Lightweight, fast model (recommended for testing)
ollama pull llama3.2:3b

# OR: Better quality model (requires more RAM)
ollama pull llama3.2:7b

# OR: Code-focused model
ollama pull qwen2.5-coder:7b
```

### 3. Verify It Works

```powershell
ollama list
```

You should see your downloaded models listed.

### 4. Restart Your GAIOL Server

The server will automatically detect Ollama on startup:

```
✅ Ollama available with 1 local models: [llama3.2:3b]
💡 Ollama will be used as backup when OpenRouter is rate-limited
```

## How It Works

When OpenRouter returns rate limits (HTTP 429), the system will automatically fall back to your local Ollama models instead of showing errors!

**Flow:**
1. Try OpenRouter (cloud)
2. If rate limited → Try Ollama (local)  
3. If Ollama unavailable → Show fallback message

## Recommended Models

| Model | Size | Best For | RAM Needed |
|-------|------|----------|------------|
| `llama3.2:1b` | ~1GB | Speed testing | 4GB |
| `llama3.2:3b` | ~2GB | General use | 8GB |
| `qwen2.5-coder:7b` | ~4GB | Coding tasks | 16GB |
| `deepseek-r1:7b` | ~5GB | Reasoning | 16GB |

## Test It

1. Submit a query to GAIOL
2. Watch the logs - you'll see:
   - First: Attempts to OpenRouter
   - If 429 errors: Falls back to Ollama
   - Response from local model!

## Troubleshooting

**"Ollama not running"**
- Make sure you ran `ollama pull <model>` first
- Check if Ollama is running: `ollama list`
- Default port is 11434

**Models too slow?**
- Use smaller models like `llama3.2:3b`
- Close other applications  
- Consider GPU acceleration (NVIDIA/AMD)

## Advanced: Custom Ollama URL

If Ollama runs on a different port/host, edit `cmd/web-server/main.go`:

```go
ollamaAdapter := adapters.NewOllamaAdapter("http://192.168.1.100:11434")
```
