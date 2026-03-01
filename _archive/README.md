# GAIOL - Go AI Orchestration Layer

<div align="center">

**A comprehensive AI service orchestration platform that provides unified access to multiple AI models through intelligent routing and advanced reasoning capabilities.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-success?style=flat-square)]()

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Frontend Features](#frontend-features)
- [Reasoning Engine](#reasoning-engine)
- [Authentication](#authentication)
- [Database Setup](#database-setup)
- [Development](#development)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

GAIOL (Go AI Orchestration Layer) is a production-ready platform that unifies access to multiple AI models and services through a standardized **UAIP (Universal AI Protocol)** interface. It provides intelligent routing, multi-model orchestration, advanced reasoning capabilities, and a modern web interface for interacting with AI models.

### What Makes GAIOL Unique?

- **🔄 Multi-Provider Support**: Seamlessly integrates with OpenRouter, Google Gemini, and HuggingFace
- **🧠 Intelligent Routing**: Automatically selects the best model based on task, cost, and quality requirements
- **⚡ Advanced Reasoning Engine**: Decomposes complex queries into steps and uses beam search to find optimal solutions
- **🎨 Modern Web Interface**: Beautiful, responsive UI with real-time reasoning visualization
- **🔐 Enterprise-Ready**: Built-in authentication, multi-tenancy, and usage tracking
- **📊 Performance Monitoring**: Real-time metrics and cost tracking

---

## ✨ Key Features

### Core Capabilities

- **Unified AI Protocol (UAIP)**: Standardized interface for all AI model interactions
- **Model Registry**: Centralized registry of 100+ AI models from multiple providers
- **Smart Routing**: Intelligent model selection based on strategy (cost, quality, speed, balanced)
- **Multi-Model Comparison**: Query multiple models simultaneously and compare results
- **Reasoning Engine**: Advanced multi-step reasoning with beam search and consensus
- **RAG Integration**: Retrieval-Augmented Generation for context-aware responses
- **Performance Tracking**: Real-time latency, cost, and quality metrics

### Frontend Features

- **Interactive Chat Interface**: Clean, modern chat UI with message history
- **Model Comparison**: Side-by-side comparison of model responses
- **Reasoning Visualization**: Real-time visualization of reasoning steps and beam search
- **Voice Input**: Speech-to-text using Web Speech API
- **File Attachments**: Upload and process text files (.txt, .md, .json, .csv)
- **Prompt Library**: Pre-built prompt templates for common tasks
- **Global Search**: Quick search across models and history (⌘K / Ctrl+K)
- **Settings Management**: Customizable defaults for strategy, tokens, and temperature
- **History Management**: View and replay previous queries
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile

### Enterprise Features

- **User Authentication**: Supabase-based authentication with JWT tokens
- **Multi-Tenancy**: Row-level security with tenant isolation
- **Usage Tracking**: Comprehensive analytics and cost tracking per tenant
- **Session Management**: Secure session handling with token refresh
- **API Rate Limiting**: Built-in protection against abuse

---

## 🏗️ Architecture

GAIOL follows a modular, service-oriented architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                      Web Frontend                            │
│  (HTML/CSS/JS - Chat, Models, Reasoning, History, Settings) │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ HTTP/WebSocket
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   Web Server (Go)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   REST API  │  │  WebSocket   │  │  Auth API     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼──────┐ ┌─────▼──────┐ ┌────▼──────┐
│   Model      │ │ Reasoning   │ │  Database │
│   Registry   │ │   Engine    │ │ (Supabase)│
└───────┬──────┘ └─────┬──────┘ └────┬──────┘
        │              │              │
┌───────▼──────────────▼──────────────▼──────┐
│         Model Adapters                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │OpenRouter│  │ Gemini   │  │HuggingFace│ │
│  └──────────┘  └──────────┘  └──────────┘ │
└───────────────────────────────────────────┘
```

### Core Components

1. **Model Registry**: Centralized catalog of available AI models
2. **Model Router**: Intelligent routing based on strategy and requirements
3. **Reasoning Engine**: Multi-step reasoning with beam search
4. **UAIP Protocol**: Standardized request/response format
5. **Authentication**: Supabase-based auth with JWT validation
6. **Database**: PostgreSQL (Supabase) for persistence and analytics

For detailed architecture documentation, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** ([Download](https://golang.org/dl/))
- **API Keys** (at least one):
  - `OPENROUTER_API_KEY` (required for most models)
  - `GEMINI_API_KEY` (optional, for Google Gemini)
  - `HUGGINGFACE_API_KEY` (optional, for HuggingFace models)

### 1. Clone and Install

```bash
git clone <repository-url>
cd GAIOL
go mod download
```

### 2. Configure Environment

Create a `.env` file in the project root:

```env
# Required
OPENROUTER_API_KEY=your-openrouter-key-here

# Optional
GEMINI_API_KEY=your-gemini-key-here
HUGGINGFACE_API_KEY=your-huggingface-key-here

# Database (optional, for auth and persistence)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here

# Server
PORT=8080
```

### 3. Start the Server

**Option A: Using Go**
```bash
go run cmd/web-server/main.go
```

**Option B: Using Make**
```bash
make run
```

**Option C: Using Scripts**
- Windows: `start.bat`
- Linux/Mac: `./start.sh`
- PowerShell: `.\start.ps1`

### 4. Access the Web Interface

Open your browser and navigate to:
```
http://localhost:8080
```

### 5. Test the API

```bash
# Health check
curl http://localhost:8080/health

# List available models
curl http://localhost:8080/api/models

# Query with smart routing
curl -X POST http://localhost:8080/api/query/smart \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Explain quantum computing in simple terms"}'
```

---

## 📦 Installation

### From Source

```bash
# Clone repository
git clone <repository-url>
cd GAIOL

# Install dependencies
go mod download

# Build
go build -o gaiol ./cmd/web-server

# Run
./gaiol
```

### Using Docker

```bash
# Build image
docker build -t gaiol:latest .

# Run container
docker run -p 8080:8080 \
  -e OPENROUTER_API_KEY=your-key \
  gaiol:latest
```

### Using Docker Compose

```bash
docker-compose -f docker-compose.dev.yml up
```

---

## ⚙️ Configuration

### Environment Variables

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| `OPENROUTER_API_KEY` | Yes | OpenRouter API key | - |
| `GEMINI_API_KEY` | No | Google Gemini API key | - |
| `HUGGINGFACE_API_KEY` | No | HuggingFace API key | - |
| `SUPABASE_URL` | No | Supabase project URL | - |
| `SUPABASE_ANON_KEY` | No | Supabase anonymous key | - |
| `PORT` | No | Server port | `8080` |

### Routing Strategies

The smart router supports four strategies:

- **`free_only`**: Only use free models
- **`lowest_cost`**: Minimize cost per token
- **`highest_quality`**: Maximize quality score
- **`balanced`**: Balance cost, quality, and speed (default)

### Reasoning Engine Configuration

```go
BeamConfig{
    Enabled:   true,
    BeamWidth: 3,  // Number of paths to explore
}

ConsensusConfig{
    Enabled:   true,
    Strategy:  "meta_agent",  // majority, weighted, or meta_agent
    MetaModel: "openrouter:google/gemini-2.0-flash-exp:free",
    Threshold: 0.6,
}
```

---

## 📚 API Documentation

GAIOL provides a comprehensive REST API and WebSocket interface. For complete API documentation, see [API.md](API.md).

### Core Endpoints

#### Model Discovery
- `GET /api/models` - List all available models
- `GET /api/models/free` - List free models
- `GET /api/models/:provider` - List models by provider

#### Query Endpoints
- `POST /api/query` - Multi-model comparison (legacy)
- `POST /api/query/smart` - Smart routing (recommended)
- `POST /api/query/model` - Query specific model by ID

#### Reasoning Engine
- `POST /api/reasoning/start` - Start reasoning session
- `GET /api/reasoning/status/:session_id` - Get session status
- `WS /api/reasoning/ws` - WebSocket for real-time updates

#### Authentication
- `POST /api/auth/signup` - Create account
- `POST /api/auth/signin` - Sign in
- `POST /api/auth/signout` - Sign out
- `GET /api/auth/session` - Get current session
- `POST /api/auth/refresh` - Refresh token

#### System
- `GET /health` - Health check

### Example Requests

**Smart Query:**
```bash
curl -X POST http://localhost:8080/api/query/smart \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Write a Python function to calculate fibonacci numbers",
    "strategy": "balanced",
    "max_tokens": 500,
    "temperature": 0.7
  }'
```

**Start Reasoning Session:**
```bash
curl -X POST http://localhost:8080/api/reasoning/start \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Design a REST API for a todo application",
    "models": []
  }'
```

---

## 🎨 Frontend Features

The GAIOL web interface provides a modern, feature-rich experience:

### Pages

- **Chat**: Main interface for querying models
- **Models**: Browse and filter available models
- **Reasoning**: Visualize reasoning engine execution
- **History**: View and replay previous queries
- **Settings**: Configure defaults and preferences
- **Profile**: User account management
- **Observability**: System metrics and monitoring

### Interactive Features

- **Voice Input**: Click microphone to speak your query
- **File Attachments**: Upload text files for processing
- **Prompt Library**: Browse pre-built prompt templates
- **Global Search**: Press ⌘K / Ctrl+K to search models and history
- **Model Comparison**: Side-by-side comparison of responses
- **Real-time Updates**: WebSocket-based live reasoning visualization

For detailed frontend documentation, see [FEATURES_IMPLEMENTED.md](FEATURES_IMPLEMENTED.md).

---

## 🧠 Reasoning Engine

The GAIOL Reasoning Engine is an advanced multi-step reasoning system that:

1. **Decomposes** complex queries into logical steps
2. **Explores** multiple reasoning paths using beam search
3. **Scores** outputs from multiple models
4. **Selects** the best path based on cumulative scores
5. **Synthesizes** final output using consensus mechanisms

### How It Works

```
User Prompt
    ↓
Decompose into Steps
    ↓
For each step:
  ├─ Run multiple models in parallel
  ├─ Score all outputs
  ├─ Use beam search (keep top N paths)
  └─ Extend paths with model outputs
    ↓
Select Best Path (highest score)
    ↓
Apply Consensus (if enabled)
    ↓
Combine Final Output
```

### Features

- **Beam Search**: Explores multiple reasoning paths simultaneously
- **Consensus Mechanisms**: Majority voting, weighted voting, or meta-agent synthesis
- **Auto-Model Selection**: Automatically selects 4 free models if none specified
- **Real-time Updates**: WebSocket events for live progress tracking
- **Database Persistence**: Saves sessions and steps for analysis

For detailed reasoning engine documentation, see [SIMPLIFIED_ARCHITECTURE.md](SIMPLIFIED_ARCHITECTURE.md).

---

## 🔐 Authentication

GAIOL uses Supabase Auth for authentication. The system supports:

- Email/password authentication
- JWT token-based sessions
- Automatic token refresh
- Multi-tenant user isolation
- Row-level security (RLS)

### Setup

1. Create a Supabase project at [supabase.com](https://supabase.com)
2. Add credentials to `.env`:
   ```env
   SUPABASE_URL=https://your-project.supabase.co
   SUPABASE_ANON_KEY=your-anon-key
   ```
3. Run database migrations (see [DATABASE_SETUP.md](DATABASE_SETUP.md))

### Usage

Authentication is optional - the system works without a database, but features like history persistence and user-specific settings require authentication.

For complete authentication documentation, see [AUTHENTICATION.md](AUTHENTICATION.md).

---

## 🗄️ Database Setup

GAIOL uses Supabase (PostgreSQL) for:

- User authentication and profiles
- Multi-tenant data isolation
- Query history and analytics
- Performance tracking
- Session persistence

### Quick Setup

1. Create Supabase project
2. Run migrations from `migrations/` directory
3. Configure environment variables
4. Start server - database connection is automatic

For detailed setup instructions, see [DATABASE_SETUP.md](DATABASE_SETUP.md).

---

## 🛠️ Development

### Project Structure

```
GAIOL/
├── cmd/                    # Main applications
│   ├── web-server/         # Web server (main entry point)
│   ├── uaip-service/       # Standalone UAIP service
│   └── test-*/            # Test utilities
├── internal/              # Private application code
│   ├── auth/              # Authentication
│   ├── database/          # Database client
│   ├── models/            # Model registry and routing
│   │   └── adapters/     # Provider adapters
│   ├── reasoning/         # Reasoning engine
│   ├── monitoring/       # Metrics and monitoring
│   └── uaip/              # UAIP protocol
├── web/                   # Frontend
│   ├── css/              # Stylesheets
│   ├── js/               # JavaScript modules
│   └── *.html            # HTML pages
├── migrations/           # Database migrations
└── Makefile              # Build automation
```

### Building

```bash
# Build all binaries
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
make coverage

# Test specific adapter
go run cmd/test-openrouter/main.go
```

### Code Style

- Follow Go standard formatting (`gofmt`)
- Use `golangci-lint` for linting
- Write tests for new features
- Document public APIs

---

## 📊 Project Structure

### Backend Components

| Component | Location | Description |
|-----------|----------|-------------|
| Web Server | `cmd/web-server/` | Main HTTP server and API |
| Model Registry | `internal/models/registry.go` | Model catalog management |
| Model Router | `internal/models/router.go` | Intelligent routing logic |
| Reasoning Engine | `internal/reasoning/` | Multi-step reasoning system |
| UAIP Protocol | `internal/uaip/` | Standardized protocol |
| Authentication | `internal/auth/` | Auth API and middleware |
| Database | `internal/database/` | Supabase client and helpers |

### Frontend Components

| Component | Location | Description |
|-----------|----------|-------------|
| Main UI | `web/index.html` | Chat interface |
| API Client | `web/js/api.js` | API communication |
| State Management | `web/js/state.js` | Application state |
| Reasoning UI | `web/js/reasoning-*.js` | Reasoning visualization |
| Navigation | `web/js/navigation.js` | Page routing |
| Features | `web/js/features.js` | Voice, file, search, etc. |

---

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines

- Write clear commit messages
- Add tests for new features
- Update documentation
- Follow existing code style
- Ensure all tests pass

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **OpenRouter** for providing unified access to multiple AI models
- **Google** for Gemini API
- **HuggingFace** for open-source model hosting
- **Supabase** for authentication and database infrastructure

---

## 📞 Support

- **Documentation**: See individual `.md` files in the repository
- **Issues**: Open an issue on GitHub
- **Questions**: Check existing documentation or open a discussion

---

<div align="center">

**Built with ❤️ using Go**

[⬆ Back to Top](#gaiol---go-ai-orchestration-layer)

</div>
