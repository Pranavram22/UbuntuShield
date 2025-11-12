# 🌐 Multi-Server SaaS Architecture Design

## 📊 Architecture Comparison

### Option 1: Agent-Based Architecture (⭐ RECOMMENDED)
```
┌─────────────────────────────────────────────────────────────┐
│                    CENTRAL DASHBOARD                        │
│              (Web UI + API Server)                          │
│         https://dashboard.yourcompany.com                   │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ HTTPS/WebSocket
                     │
        ┌────────────┼────────────┬─────────────────┐
        │            │            │                 │
   ┌────▼────┐  ┌───▼─────┐  ┌──▼──────┐    ┌────▼─────┐
   │ Agent 1 │  │ Agent 2 │  │ Agent 3 │    │ Agent N  │
   │ Server1 │  │ Server2 │  │ Server3 │    │ Server N │
   └─────────┘  └─────────┘  └─────────┘    └──────────┘
   
   Each agent:
   • Runs Lynis locally
   • Collects metrics
   • Sends to central dashboard
   • Lightweight (~10 MB RAM)
```

**Pros:**
- ✅ Scalable to 1000+ servers
- ✅ Real-time monitoring
- ✅ Secure (agents push data)
- ✅ Easy deployment
- ✅ Works with firewalls
- ✅ Low resource usage

**Cons:**
- ⚠️ Requires agent installation on each server
- ⚠️ Need central database

---

### Option 2: Agentless SSH-Based (Not Recommended for SaaS)
```
┌──────────────────────────────────────┐
│        CENTRAL DASHBOARD             │
│  • Stores SSH credentials            │
│  • Connects via SSH                  │
└──────────┬───────────────────────────┘
           │
           │ SSH Connections
           │
    ┌──────┼──────┬─────────┐
    │      │      │         │
┌───▼──┐ ┌─▼───┐ ┌▼────┐ ┌─▼────┐
│Srv 1 │ │Srv 2│ │Srv 3│ │Srv N │
└──────┘ └─────┘ └─────┘ └──────┘
```

**Pros:**
- ✅ No agent needed

**Cons:**
- ❌ Security risk (storing credentials)
- ❌ Doesn't scale well
- ❌ SSH connection overhead
- ❌ Firewall issues
- ❌ Not suitable for SaaS

---

### Option 3: Hybrid Pull-Push (Good for Enterprise)
```
┌────────────────────────────────────────────────┐
│           CENTRAL DASHBOARD                    │
│  • REST API for agents                         │
│  • Optional SSH fallback                       │
└────────────────┬───────────────────────────────┘
                 │
        ┌────────┼──────────┐
        │        │          │
   ┌────▼─┐  ┌──▼───┐  ┌───▼──┐
   │Agent │  │Agent │  │Agent │
   │(Push)│  │(Push)│  │(Pull)│
   └──────┘  └──────┘  └──────┘
```

---

## 🎯 RECOMMENDED: Agent-Based Architecture

### Why This is Best for Open Source SaaS:

1. **Scalability** - Can handle thousands of servers
2. **Security** - No credentials stored centrally
3. **Performance** - Lightweight agents
4. **Flexibility** - Self-hosted or cloud
5. **Open Source Friendly** - Easy to audit and contribute
6. **Modern** - Similar to Datadog, New Relic, Prometheus

---

## 🏗️ Complete System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    INTERNET / PRIVATE NETWORK                   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │
┌───────────────────────────────┼─────────────────────────────────┐
│                    CENTRAL PLATFORM                             │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Frontend (React/Vue/Svelte)                 │  │
│  │  • Multi-server dashboard                                │  │
│  │  • Real-time charts                                      │  │
│  │  • Alert management                                      │  │
│  └────────────────────────┬─────────────────────────────────┘  │
│                           │                                     │
│  ┌────────────────────────▼─────────────────────────────────┐  │
│  │            API Server (Go)                               │  │
│  │  Endpoints:                                              │  │
│  │  • POST /api/agents/register                            │  │
│  │  │  POST /api/agents/heartbeat                            │  │
│  │  • POST /api/metrics                                    │  │
│  │  • GET  /api/servers                                    │  │
│  │  • GET  /api/servers/:id/metrics                        │  │
│  │  • GET  /api/alerts                                     │  │
│  └────────────────────────┬─────────────────────────────────┘  │
│                           │                                     │
│  ┌────────────────────────▼─────────────────────────────────┐  │
│  │            Database (PostgreSQL + TimescaleDB)           │  │
│  │  Tables:                                                 │  │
│  │  • servers (id, name, ip, agent_version, status)       │  │
│  │  • metrics (server_id, timestamp, data)                 │  │
│  │  • audits (server_id, timestamp, results)               │  │
│  │  • alerts (server_id, type, severity, timestamp)        │  │
│  │  • users (id, email, org_id)                           │  │
│  │  • organizations (id, name, plan)                       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            Cache Layer (Redis)                           │  │
│  │  • Real-time metrics                                     │  │
│  │  • Session management                                    │  │
│  │  • Rate limiting                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            Message Queue (Redis/RabbitMQ)                │  │
│  │  • Agent registration events                             │  │
│  │  • Alert notifications                                   │  │
│  │  • Background jobs                                       │  │
│  └──────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               │ HTTPS/WebSocket
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼────────┐    ┌────────▼───────┐    ┌───────▼────────┐
│  Server 1      │    │  Server 2      │    │  Server N      │
│  ┌──────────┐  │    │  ┌──────────┐  │    │  ┌──────────┐  │
│  │  Agent   │  │    │  │  Agent   │  │    │  │  Agent   │  │
│  │          │  │    │  │          │  │    │  │          │  │
│  │ • Lynis  │  │    │  │ • Lynis  │  │    │  │ • Lynis  │  │
│  │ • Metrics│  │    │  │ • Metrics│  │    │  │ • Metrics│  │
│  │ • Push   │  │    │  │ • Push   │  │    │  │ • Push   │  │
│  └──────────┘  │    │  └──────────┘  │    │  └──────────┘  │
└────────────────┘    └────────────────┘    └────────────────┘
```

---

## 📦 Components Breakdown

### 1. Central Dashboard (Go Backend)
- REST API for agents
- WebSocket for real-time updates
- User authentication (JWT)
- Multi-tenancy support
- API rate limiting

### 2. Agent (Go - Lightweight)
- Runs on each monitored server
- Executes Lynis audits
- Collects system metrics
- Sends data to central dashboard
- Self-updating capability
- ~10 MB RAM usage

### 3. Database
- **PostgreSQL** - Main data store
- **TimescaleDB** - Time-series metrics
- **Redis** - Caching & sessions

### 4. Frontend
- React/Vue/Svelte dashboard
- Real-time charts (Chart.js/D3.js)
- Multi-server overview
- Alert management

---

## 🔐 Security Architecture

```
┌─────────────────────────────────────────┐
│        Security Layers                  │
├─────────────────────────────────────────┤
│                                         │
│  1. Agent Authentication                │
│     • API Key per server                │
│     • JWT tokens                        │
│     • Certificate-based auth (optional) │
│                                         │
│  2. Data Encryption                     │
│     • TLS 1.3 in transit                │
│     • AES-256 at rest                   │
│                                         │
│  3. Multi-Tenancy Isolation             │
│     • Organization-based separation     │
│     • Row-level security (RLS)          │
│                                         │
│  4. API Security                        │
│     • Rate limiting                     │
│     • CORS policies                     │
│     • Input validation                  │
│                                         │
│  5. Audit Logging                       │
│     • All API calls logged              │
│     • Agent activity tracked            │
│                                         │
└─────────────────────────────────────────┘
```

---

## 📊 Data Flow

### Agent Registration Flow
```
1. User installs agent on server
   ↓
2. Agent generates unique ID
   ↓
3. Agent calls: POST /api/agents/register
   Body: { hostname, ip, os, version }
   ↓
4. Dashboard returns API key
   ↓
5. Agent stores API key locally
   ↓
6. Agent starts sending heartbeats
```

### Audit Data Flow
```
1. Agent runs Lynis (scheduled)
   ↓
2. Agent parses results
   ↓
3. Agent sends: POST /api/metrics
   Headers: Authorization: Bearer <api_key>
   Body: { audit_data, timestamp }
   ↓
4. Dashboard validates & stores
   ↓
5. Dashboard checks for alerts
   ↓
6. Dashboard updates real-time UI
```

### Real-Time Updates Flow
```
1. User opens dashboard
   ↓
2. Frontend connects via WebSocket
   ↓
3. Agent sends new metrics
   ↓
4. Backend processes metrics
   ↓
5. Backend pushes to WebSocket
   ↓
6. Frontend updates charts in real-time
```

---

## 💾 Database Schema

```sql
-- Organizations (Multi-tenancy)
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    max_servers INT DEFAULT 5,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    org_id UUID REFERENCES organizations(id),
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Servers (Monitored systems)
CREATE TABLE servers (
    id UUID PRIMARY KEY,
    org_id UUID REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    ip_address INET,
    os VARCHAR(100),
    agent_version VARCHAR(50),
    api_key VARCHAR(255) UNIQUE,
    status VARCHAR(50) DEFAULT 'active',
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_org_id (org_id),
    INDEX idx_status (status)
);

-- Audit Results (Historical data)
CREATE TABLE audits (
    id UUID PRIMARY KEY,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    hardening_index INT,
    warnings INT,
    tests_performed INT,
    compliance_scores JSONB,
    raw_data JSONB,
    INDEX idx_server_timestamp (server_id, timestamp DESC)
);

-- Metrics (Time-series data using TimescaleDB)
CREATE TABLE metrics (
    timestamp TIMESTAMP NOT NULL,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL,
    metric_value FLOAT NOT NULL,
    metadata JSONB
);

-- Convert to hypertable for TimescaleDB
SELECT create_hypertable('metrics', 'timestamp');

-- Alerts
CREATE TABLE alerts (
    id UUID PRIMARY KEY,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    severity VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP,
    INDEX idx_server_status (server_id, status)
);

-- Agent Activity Log
CREATE TABLE agent_logs (
    id BIGSERIAL PRIMARY KEY,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    timestamp TIMESTAMP DEFAULT NOW(),
    INDEX idx_server_timestamp (server_id, timestamp DESC)
);
```

---

## 🚀 Deployment Models

### 1. SaaS Cloud (Managed)
```
• You host central dashboard
• Users install agents on their servers
• Agents connect to your cloud
• Pricing: Per server/month
```

### 2. Self-Hosted (Open Source)
```
• Users deploy entire stack
• Full control over data
• Run on their infrastructure
• Free (open source)
```

### 3. Hybrid
```
• Dashboard can be cloud or self-hosted
• Agents always on customer servers
• Flexible deployment
```

---

## 📈 Scalability Considerations

### For 100 Servers
- Single server deployment
- PostgreSQL + Redis on same machine
- ~4 GB RAM, 2 CPU cores

### For 1,000 Servers
- Separate DB and API servers
- PostgreSQL with read replicas
- Redis cluster for caching
- ~16 GB RAM, 4-8 CPU cores

### For 10,000+ Servers
- Kubernetes deployment
- Horizontal auto-scaling
- Multiple DB replicas
- Distributed caching
- Load balancers

---

## 💰 Pricing Model (SaaS)

```
Free Tier:
• Up to 5 servers
• 7 days data retention
• Basic alerts

Pro Tier ($29/month):
• Up to 50 servers
• 90 days data retention
• Advanced alerts
• Email notifications

Enterprise Tier ($299/month):
• Unlimited servers
• 1 year data retention
• Custom alerts
• Slack/Webhook integration
• SSO support
• Priority support
```

---

## 🎯 Implementation Phases

### Phase 1: Core Agent System (Week 1-2)
- [ ] Agent registration
- [ ] Agent heartbeat
- [ ] Basic metric collection
- [ ] Central API server

### Phase 2: Dashboard UI (Week 3-4)
- [ ] Multi-server view
- [ ] Real-time updates
- [ ] Basic charts

### Phase 3: Advanced Features (Week 5-6)
- [ ] Alerts system
- [ ] Historical trends
- [ ] Comparison views

### Phase 4: Multi-Tenancy (Week 7-8)
- [ ] User authentication
- [ ] Organization management
- [ ] Access control

### Phase 5: Production Ready (Week 9-10)
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Documentation
- [ ] Deployment automation

---

## 🔧 Technology Stack Recommendation

### Backend
- **Language**: Go (current)
- **Framework**: Gin or Echo
- **Database**: PostgreSQL + TimescaleDB
- **Cache**: Redis
- **WebSocket**: gorilla/websocket

### Agent
- **Language**: Go
- **Size**: ~15 MB binary
- **Dependencies**: None (static binary)

### Frontend
- **Framework**: React (or Vue/Svelte)
- **Charts**: Chart.js or Apache ECharts
- **State**: Redux or Zustand
- **Real-time**: WebSocket

### DevOps
- **Containers**: Docker
- **Orchestration**: Docker Compose (simple) or Kubernetes (scale)
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana (for the platform itself)

---

## 🎨 Dashboard UI Mockup

```
┌────────────────────────────────────────────────────────────┐
│ UbuntuShield Dashboard    [+ Add Server]  [Alerts: 3]  👤 │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Overview                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │ 25       │ │ 87%      │ │ 12       │ │ 98.5%    │    │
│  │ Servers  │ │ Avg Score│ │ Alerts   │ │ Uptime   │    │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
│                                                            │
│  Servers                              [🔍 Search] [Filter]│
│  ┌────────────────────────────────────────────────────┐  │
│  │ Name          IP           Score  Status  Last Seen │  │
│  ├────────────────────────────────────────────────────┤  │
│  │ ● prod-web-01 10.0.1.10   92%   ✓ Active  2m ago   │  │
│  │ ● prod-web-02 10.0.1.11   88%   ✓ Active  1m ago   │  │
│  │ ● prod-db-01  10.0.2.10   95%   ✓ Active  30s ago  │  │
│  │ ⚠ staging-01  10.0.3.10   65%   ! Warning 5m ago   │  │
│  │ ● dev-01      10.0.4.10   78%   ✓ Active  1m ago   │  │
│  └────────────────────────────────────────────────────┘  │
│                                                            │
│  Security Trends (Last 30 Days)                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │    [Line chart showing security scores over time]  │  │
│  │     All servers averaged                            │  │
│  └────────────────────────────────────────────────────┘  │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 🎯 RECOMMENDATION

**Use Agent-Based Architecture with the following tech stack:**

1. **Backend**: Go (your current codebase as base)
2. **Agent**: Go (lightweight, single binary)
3. **Database**: PostgreSQL + TimescaleDB
4. **Frontend**: React with real-time updates
5. **Deployment**: Docker Compose (start) → Kubernetes (scale)

This gives you:
- ✅ Scalability to 10,000+ servers
- ✅ Open source friendly
- ✅ Self-hosted or SaaS options
- ✅ Modern architecture
- ✅ Easy to maintain

---

**Ready to implement? I can build this step by step!**

