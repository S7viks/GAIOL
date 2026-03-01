# GAIOL Architecture Documentation

Comprehensive architecture overview of the GAIOL platform.

---

## Table of Contents

- [System Overview](#system-overview)
- [Architecture Layers](#architecture-layers)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Reasoning Engine](#reasoning-engine)
- [Model Adapters](#model-adapters)
- [Database Schema](#database-schema)
- [Security Architecture](#security-architecture)
- [Deployment Architecture](#deployment-architecture)

---

## System Overview

GAIOL is a multi-layered system that provides unified access to AI models through intelligent orchestration and advanced reasoning capabilities.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Web Browser  │  │  Mobile App  │  │  API Client  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTP/WebSocket
┌──────────────────────▼──────────────────────────────────────┐
│                  Application Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  REST API    │  │  WebSocket   │  │  Auth API     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼──────┐ ┌─────▼──────┐ ┌────▼──────┐
│   Business   │ │ Reasoning   │ │  Data     │
│   Logic      │ │   Engine    │ │  Layer    │
└───────┬──────┘ └─────┬──────┘ └────┬──────┘
        │              │              │
┌───────▼──────────────▼──────────────▼──────┐
│            Service Layer                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐│
│  │ Registry │  │  Router   │  │ Adapters ││
│  └──────────┘  └──────────┘  └──────────┘│
└───────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│            External Services                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐│
│  │OpenRouter│  │  Gemini  │  │HuggingFace││
│  └──────────┘  └──────────┘  └──────────┘│
└─────────────────────────────────────────────┘
```

---

## Architecture Layers

### 1. Presentation Layer

**Components:**
- Web frontend (HTML/CSS/JavaScript)
- REST API endpoints
- WebSocket connections

**Responsibilities:**
- User interface rendering
- Request/response handling
- Real-time updates
- Authentication UI

### 2. Application Layer

**Components:**
- HTTP handlers
- WebSocket handlers
- Authentication middleware
- Request validation

**Responsibilities:**
- Request routing
- Authentication/authorization
- Input validation
- Response formatting

### 3. Business Logic Layer

**Components:**
- Model Registry
- Model Router
- Reasoning Engine
- Performance Tracker

**Responsibilities:**
- Business rules
- Model selection logic
- Reasoning orchestration
- Performance tracking

### 4. Service Layer

**Components:**
- Model Adapters
- UAIP Protocol
- Database Client
- Monitoring Service

**Responsibilities:**
- External API communication
- Protocol standardization
- Data persistence
- Metrics collection

### 5. Infrastructure Layer

**Components:**
- Supabase (Database)
- External AI Providers
- File System
- Environment Configuration

**Responsibilities:**
- Data storage
- External service integration
- Configuration management

---

## Core Components

### 1. Model Registry

**Location:** `internal/models/registry.go`

**Purpose:** Centralized catalog of all available AI models.

**Key Features:**
- Model discovery and registration
- Provider abstraction
- Metadata management
- Capability tracking

**Data Structures:**
```go
type ModelMetadata struct {
    ID            ModelID
    Provider      string
    ModelName     string
    DisplayName   string
    CostInfo      CostInfo
    Capabilities  ModelCapabilities
    QualityScore  float64
    ContextWindow int
    MaxTokens     int
    Tags          []string
    Adapter       ModelAdapter
}
```

### 2. Model Router

**Location:** `internal/models/router.go`

**Purpose:** Intelligent model selection based on strategy and requirements.

**Routing Strategies:**
- `free_only`: Only free models
- `lowest_cost`: Minimize cost
- `highest_quality`: Maximize quality
- `balanced`: Balance all factors (default)

**Selection Algorithm:**
1. Filter models by requirements
2. Score each model based on strategy
3. Select top N models
4. Return ranked list

### 3. Reasoning Engine

**Location:** `internal/reasoning/`

**Purpose:** Multi-step reasoning with beam search and consensus.

**Components:**
- **Decomposer**: Breaks queries into steps
- **Orchestrator**: Runs models in parallel
- **Scorer**: Evaluates outputs
- **Beam Search**: Explores multiple paths
- **Consensus**: Synthesizes final output

**Flow:**
```
Prompt → Decompose → Steps → For Each Step:
  ├─ Run Models (Parallel)
  ├─ Score Outputs
  ├─ Beam Search (Keep Top N)
  └─ Extend Paths
→ Select Best Path → Consensus → Final Output
```

### 4. UAIP Protocol

**Location:** `internal/uaip/`

**Purpose:** Standardized protocol for AI model communication.

**Request Structure:**
```go
type UAIPRequest struct {
    UAIP   UAIPHeader
    Payload Payload
}

type UAIPHeader struct {
    Version   string
    MessageID string
    Timestamp time.Time
}

type Payload struct {
    Input              PayloadInput
    OutputRequirements OutputRequirements
}
```

**Response Structure:**
```go
type UAIPResponse struct {
    UAIP     UAIPHeader
    Result   Result
    Metadata Metadata
}
```

### 5. Authentication System

**Location:** `internal/auth/`

**Purpose:** User authentication and authorization.

**Components:**
- Supabase Auth integration
- JWT token validation
- Session management
- Multi-tenant isolation

**Flow:**
```
User → Sign In → Supabase Auth → JWT Token → 
  Validate Token → Extract User/Tenant → 
  Add to Context → RLS Filtering
```

---

## Data Flow

### Query Flow

```
1. User submits query
   ↓
2. Frontend sends POST /api/query/smart
   ↓
3. Handler validates request
   ↓
4. Router selects models based on strategy
   ↓
5. Reasoning Engine processes:
   a. Decompose prompt
   b. For each step:
      - Run models in parallel
      - Score outputs
      - Beam search
   c. Select best path
   d. Apply consensus
   ↓
6. Return final output
   ↓
7. Frontend displays result
```

### Authentication Flow

```
1. User submits credentials
   ↓
2. POST /api/auth/signin
   ↓
3. Supabase Auth validates
   ↓
4. Return JWT tokens
   ↓
5. Client stores tokens
   ↓
6. Subsequent requests include token
   ↓
7. Middleware validates token
   ↓
8. Extract user/tenant context
   ↓
9. Apply RLS filtering
```

### Reasoning Session Flow

```
1. POST /api/reasoning/start
   ↓
2. Create session
   ↓
3. Decompose prompt
   ↓
4. WebSocket: decompose_start
   ↓
5. For each step:
   a. WebSocket: step_start
   b. Run models in parallel
   c. WebSocket: model_response (for each)
   d. Score outputs
   e. Beam search
   f. WebSocket: beam_update
   g. Consensus (if enabled)
   h. WebSocket: consensus
   i. WebSocket: step_end
   ↓
6. Select best path
   ↓
7. WebSocket: reasoning_end
   ↓
8. Save session to database
```

---

## Reasoning Engine

### Architecture

```
┌─────────────────────────────────────────┐
│         Reasoning Engine                │
├─────────────────────────────────────────┤
│  ┌──────────────┐                       │
│  │ Decomposer   │ → Breaks into steps   │
│  └──────┬───────┘                       │
│         │                                │
│  ┌──────▼───────┐                       │
│  │ Orchestrator │ → Runs models         │
│  └──────┬───────┘                       │
│         │                                │
│  ┌──────▼───────┐                       │
│  │   Scorer     │ → Evaluates outputs   │
│  └──────┬───────┘                       │
│         │                                │
│  ┌──────▼───────┐                       │
│  │ Beam Search  │ → Explores paths      │
│  └──────┬───────┘                       │
│         │                                │
│  ┌──────▼───────┐                       │
│  │  Consensus   │ → Synthesizes result  │
│  └──────────────┘                       │
└─────────────────────────────────────────┘
```

### Beam Search Algorithm

1. **Initialization**: Start with empty path
2. **Expansion**: For each active path, run all models
3. **Scoring**: Score all new paths
4. **Pruning**: Keep top N paths (beam width)
5. **Selection**: Choose path with highest cumulative score

### Consensus Mechanisms

1. **Majority Voting**: Simple vote counting
2. **Weighted Voting**: Votes weighted by scores
3. **Meta-Agent**: LLM synthesizes best answer (default)

---

## Model Adapters

### Adapter Interface

```go
type ModelAdapter interface {
    Name() string
    Provider() string
    SupportedTasks() []TaskType
    RequiresAuth() bool
    GetCapabilities() ModelCapabilities
    GetCost() CostInfo
    HealthCheck() error
    GenerateText(ctx context.Context, modelName string, req *UAIPRequest) (*UAIPResponse, error)
}
```

### Implemented Adapters

1. **OpenRouter Adapter**
   - Supports 100+ models
   - Unified API interface
   - Cost tracking

2. **Gemini Adapter**
   - Direct Google API integration
   - Multimodal support
   - Fast responses

3. **HuggingFace Adapter**
   - Open-source models
   - Inference API
   - Custom model support

---

## Database Schema

### Core Tables

#### `user_profiles`
```sql
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES auth.users(id),
    tenant_id UUID NOT NULL,
    org_id UUID,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `organizations`
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `api_queries`
```sql
CREATE TABLE api_queries (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID,
    model_id TEXT,
    prompt TEXT,
    response TEXT,
    tokens_used INTEGER,
    cost DECIMAL,
    latency_ms INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `reasoning_sessions`
```sql
CREATE TABLE reasoning_sessions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID,
    prompt TEXT,
    status TEXT,
    final_output TEXT,
    total_cost DECIMAL,
    total_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `reasoning_steps`
```sql
CREATE TABLE reasoning_steps (
    id UUID PRIMARY KEY,
    session_id UUID REFERENCES reasoning_sessions(id),
    step_id INTEGER,
    title TEXT,
    objective TEXT,
    task_type TEXT,
    status TEXT,
    final_output TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Row-Level Security (RLS)

All tables have RLS enabled with policies that:
- Filter by `tenant_id` automatically
- Allow users to see only their tenant's data
- Enforce authentication requirements

---

## Security Architecture

### Authentication

- **JWT Tokens**: Signed by Supabase
- **Token Validation**: On every protected request
- **Token Refresh**: Automatic via refresh tokens
- **Session Management**: Stateless (JWT-based)

### Authorization

- **Multi-Tenancy**: Tenant isolation via RLS
- **Row-Level Security**: Database-level filtering
- **Context Extraction**: User/tenant from JWT claims

### Data Protection

- **Encryption**: HTTPS for all communications
- **Secrets Management**: Environment variables
- **Input Validation**: All inputs validated
- **SQL Injection**: Parameterized queries only

### API Security

- **CORS**: Configurable origins
- **Rate Limiting**: (Planned for future)
- **Request Validation**: Schema validation
- **Error Handling**: No sensitive data in errors

---

## Deployment Architecture

### Development

```
Developer Machine
  ├─ Go Runtime
  ├─ Web Server (localhost:8080)
  ├─ Local .env file
  └─ Direct API calls to providers
```

### Production

```
┌─────────────────────────────────────┐
│         Load Balancer               │
└──────────────┬──────────────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
┌───▼───┐  ┌───▼───┐  ┌───▼───┐
│ App 1 │  │ App 2 │  │ App 3 │
└───┬───┘  └───┬───┘  └───┬───┘
    │          │          │
    └──────────┼──────────┘
               │
    ┌──────────▼──────────┐
    │   Supabase (DB)      │
    └──────────────────────┘
               │
    ┌──────────▼──────────┐
    │  External APIs      │
    │  (OpenRouter, etc.) │
    └──────────────────────┘
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o gaiol ./cmd/web-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gaiol .
EXPOSE 8080
CMD ["./gaiol"]
```

---

## Performance Considerations

### Optimization Strategies

1. **Parallel Execution**: Models run concurrently
2. **Caching**: Model metadata cached
3. **Connection Pooling**: Database connections pooled
4. **Async Processing**: Background tasks for non-critical operations

### Scalability

- **Stateless Design**: Horizontal scaling possible
- **Database**: Supabase handles scaling
- **External APIs**: Rate limits managed per provider

### Monitoring

- **Metrics**: Request counts, latency, errors
- **Cost Tracking**: Per-query cost calculation
- **Health Checks**: Regular service health monitoring

---

## Future Enhancements

### Planned Features

1. **Caching Layer**: Redis for response caching
2. **Message Queue**: For async processing
3. **Rate Limiting**: Per-user/tenant limits
4. **Analytics Dashboard**: Advanced metrics visualization
5. **Plugin System**: Custom adapter support

### Architecture Improvements

1. **Microservices**: Split into separate services
2. **Event-Driven**: Event sourcing for reasoning
3. **GraphQL API**: Alternative to REST
4. **gRPC**: High-performance internal communication

---

For implementation details, see the source code in `internal/` directory.
