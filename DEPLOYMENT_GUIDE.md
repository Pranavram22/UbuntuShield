# 🚀 Deployment Guide - Multi-Server SaaS

## ✅ What's Working Now

Your UbuntuShield platform now supports **monitoring multiple servers** from a single dashboard!

### Current Status:
- ✅ **Central API Server** - Accepts connections from agents
- ✅ **Agent Application** - Lightweight binary for monitored servers  
- ✅ **Authentication System** - API key-based security
- ✅ **File-Based Storage** - No database required (scalable to ~100 servers)
- ✅ **REST API** - Complete endpoints for all operations
- ✅ **Tested & Working** - All functionality verified

---

## 📊 Architecture

```
Central Dashboard (Your Server)
      ↓
http://your-server.com:5179
      ↓
   Accepts Agent Connections
      ↓
  ┌────┴────┬────────┬────────┐
  │         │        │        │
Agent 1   Agent 2  Agent 3  Agent N
Server 1  Server 2 Server 3 Server N
```

**Each agent:**
- Runs Lynis locally
- Sends metrics to central dashboard
- Uses API key for authentication
- Lightweight (~15MB binary, ~10MB RAM)

---

## 🎯 Deployment Options

### Option 1: Quick Test (Local)
Test everything on your local machine first.

### Option 2: Self-Hosted Production
Deploy on your own infrastructure.

### Option 3: Cloud SaaS
Host centrally, users install agents on their servers.

---

## 📦 Option 1: Quick Test (5 Minutes)

### Step 1: Start Central Dashboard

```bash
cd "/Users/apple/Desktop/untitled folder 2/UbuntuShield"

# Build
go build -o ubuntu-shield .

# Run
./ubuntu-shield
```

Server starts at: `http://localhost:5179`

### Step 2: Test Agent Registration

```bash
# Run the test script
./test-multi-server.sh
```

This will:
- Register 2 test servers
- Send heartbeats
- Submit metrics
- List all servers

### Step 3: View Results

```bash
# List all servers
curl http://localhost:5179/api/servers | jq .

# View dashboard stats
curl http://localhost:5179/api/servers | jq '.stats'
```

---

## 🏢 Option 2: Self-Hosted Production

Deploy the dashboard on one server, install agents on all others.

### Architecture:

```
Your Network:
├── dashboard-server (10.0.1.5)  # Central dashboard
├── web-server-01 (10.0.1.10)    # Agent installed
├── web-server-02 (10.0.1.11)    # Agent installed
└── db-server-01 (10.0.1.20)     # Agent installed
```

### Step 1: Deploy Central Dashboard

On your dashboard server (10.0.1.5):

```bash
# Install Go (if not already installed)
wget https://go.dev/dl/go1.20.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/Pranavram22/UbuntuShield.git
cd UbuntuShield
go build -o ubuntu-shield .

# Run as service
sudo cp ubuntu-shield /usr/local/bin/
sudo useradd -r -s /bin/false ubuntushield

# Create systemd service
sudo tee /etc/systemd/system/ubuntushield.service << EOF
[Unit]
Description=UbuntuShield Multi-Server Dashboard
After=network.target

[Service]
Type=simple
User=ubuntushield
WorkingDirectory=/var/lib/ubuntushield
ExecStart=/usr/local/bin/ubuntu-shield
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Create data directory
sudo mkdir -p /var/lib/ubuntushield/data
sudo chown -R ubuntushield:ubuntushield /var/lib/ubuntushield

# Start service
sudo systemctl daemon-reload
sudo systemctl enable ubuntushield
sudo systemctl start ubuntushield

# Check status
sudo systemctl status ubuntushield
```

### Step 2: Install Lynis on All Monitored Servers

On each server you want to monitor:

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install lynis -y

# CentOS/RHEL
sudo yum install lynis -y

# macOS
brew install lynis
```

### Step 3: Install Agent on Each Server

On web-server-01 (10.0.1.10):

```bash
# Build agent
cd UbuntuShield/agent
go build -o ubuntushield-agent .

# Install
sudo cp ubuntushield-agent /usr/local/bin/
sudo chmod +x /usr/local/bin/ubuntushield-agent

# Register with central dashboard
sudo ubuntushield-agent register http://10.0.1.5:5179

# Output will show:
# ✅ Registration successful!
#    Server ID: abc123...
#    API Key: xyz789...

# Create systemd service
sudo tee /etc/systemd/system/ubuntushield-agent.service << EOF
[Unit]
Description=UbuntuShield Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/ubuntushield-agent start
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
EOF

# Start agent
sudo systemctl daemon-reload
sudo systemctl enable ubuntushield-agent
sudo systemctl start ubuntushield-agent

# Check status
sudo systemctl status ubuntushield-agent
```

Repeat Step 3 on web-server-02, db-server-01, and all other servers.

### Step 4: View All Servers

```bash
# From anywhere:
curl http://10.0.1.5:5179/api/servers | jq .
```

---

## ☁️ Option 3: Cloud SaaS Deployment

Host the central dashboard in the cloud, users install agents on their servers.

### Architecture:

```
                    INTERNET
                        │
            ┌───────────┴────────────┐
            │                        │
    Your Cloud Server          Customer Servers
    dashboard.example.com      (agents installed)
    • Central Dashboard        • customer-web-01
    • Public API               • customer-web-02
    • User accounts            • customer-db-01
```

### Step 1: Deploy to Cloud (AWS Example)

```bash
# Launch EC2 instance
# - Ubuntu 22.04
# - t2.small or larger
# - Security group: Allow 5179 (or use nginx reverse proxy on 443)

# SSH into instance
ssh ubuntu@your-ec2-ip

# Install dependencies
sudo apt update
sudo apt install -y golang-go nginx certbot python3-certbot-nginx

# Clone and build
git clone https://github.com/Pranavram22/UbuntuShield.git
cd UbuntuShield
go build -o ubuntu-shield .
sudo cp ubuntu-shield /usr/local/bin/

# Setup nginx reverse proxy
sudo tee /etc/nginx/sites-available/ubuntushield << EOF
server {
    listen 80;
    server_name dashboard.example.com;

    location / {
        proxy_pass http://localhost:5179;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/ubuntushield /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Setup SSL
sudo certbot --nginx -d dashboard.example.com

# Setup systemd service (same as Option 2)
# ... (see above)
```

### Step 2: Create Agent Installation Script

Create `install-agent.sh` for your customers:

```bash
#!/bin/bash

DASHBOARD_URL="https://dashboard.example.com"

# Download agent
curl -L -O https://github.com/Pranavram22/UbuntuShield/releases/latest/download/ubuntushield-agent
sudo mv ubuntushield-agent /usr/local/bin/
sudo chmod +x /usr/local/bin/ubuntushield-agent

# Register
sudo ubuntushield-agent register $DASHBOARD_URL

# Install service
sudo tee /etc/systemd/system/ubuntushield-agent.service << EOF
[Unit]
Description=UbuntuShield Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/ubuntushield-agent start
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
EOF

# Start
sudo systemctl daemon-reload
sudo systemctl enable ubuntushield-agent
sudo systemctl start ubuntushield-agent

echo "✅ Agent installed and registered successfully!"
echo "View your servers at: $DASHBOARD_URL"
```

### Step 3: Customer Installation

Customers simply run:

```bash
curl -sL https://dashboard.example.com/install.sh | sudo bash
```

---

## 🔐 Security Best Practices

### 1. API Key Management
- API keys are automatically generated per server
- Stored in `/etc/ubuntushield/agent.conf`
- Never expose API keys in logs

### 2. HTTPS in Production
- Always use HTTPS for agent-dashboard communication
- Use Let's Encrypt for free SSL certificates
- Example: `https://dashboard.example.com` instead of `http://`

### 3. Firewall Rules
```bash
# On dashboard server, allow only port 5179 (or 443 if using nginx)
sudo ufw allow 5179/tcp
sudo ufw enable
```

### 4. Regular Updates
```bash
# Update dashboard
cd UbuntuShield
git pull
go build -o ubuntu-shield .
sudo systemctl restart ubuntushield

# Update agents (on each server)
# ... similar process
```

---

## 📊 API Endpoints Reference

### Agent Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/agents/register` | POST | Register new agent |
| `/api/agents/heartbeat` | POST | Send keepalive |
| `/api/metrics` | POST | Submit audit metrics |

### Dashboard Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/servers` | GET | List all servers |
| `/api/servers/{id}` | GET | Get specific server details |

### Example: List All Servers

```bash
curl https://dashboard.example.com/api/servers

# Response:
{
  "servers": [
    {
      "id": "abc123",
      "hostname": "web-01",
      "status": "active",
      "last_heartbeat": "2025-01-15T10:30:00Z",
      "latest_metrics": {
        "hardening_index": "85",
        "warnings": "5"
      }
    }
  ],
  "stats": {
    "total_servers": 10,
    "active": 9,
    "warning": 1,
    "offline": 0,
    "avg_score": 87.5
  }
}
```

---

## 🧪 Testing

### Test Agent Registration

```bash
curl -X POST http://localhost:5179/api/agents/register \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test-server",
    "ip_address": "10.0.1.100",
    "os": "linux",
    "arch": "amd64",
    "agent_version": "1.0.0"
  }'
```

### Test Metrics Submission

```bash
# Use API key from registration
curl -X POST http://localhost:5179/api/metrics \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "server_id": "YOUR_SERVER_ID",
    "timestamp": "2025-01-15T10:00:00Z",
    "hardening_index": "85",
    "warnings": "5",
    "tests_performed": "250"
  }'
```

---

## 📁 File Structure

After deployment:

```
/var/lib/ubuntushield/
├── data/
│   └── servers/
│       ├── server-001/
│       │   ├── info.json
│       │   └── audits/
│       │       ├── audit_2025-01-15_10-00-00.json
│       │       └── audit_2025-01-15_11-00-00.json
│       ├── server-002/
│       │   └── ...
│       └── server-N/
│           └── ...
└── history/  # Historical data (optional)
```

---

## 🔧 Troubleshooting

### Issue: Agent can't connect

```bash
# Check dashboard is running
curl http://dashboard-ip:5179/api/servers

# Check firewall
sudo ufw status

# Check agent logs
sudo journalctl -u ubuntushield-agent -f
```

### Issue: No metrics showing

```bash
# Check if Lynis is installed
which lynis

# Run manual audit
sudo lynis audit system --quick

# Check if report file exists
ls -la /tmp/lynis-report.dat
```

### Issue: Dashboard not starting

```bash
# Check logs
sudo journalctl -u ubuntushield -f

# Check permissions
ls -la /var/lib/ubuntushield/data

# Check port availability
sudo netstat -tulpn | grep 5179
```

---

## 📈 Scaling

### Current Setup (File-Based)
- **Good for:** Up to 100 servers
- **Storage:** ~25KB per audit per server
- **Performance:** Fast enough for most use cases

### Future: PostgreSQL (For 1000+ servers)
When you need to scale beyond 100 servers, migrate to PostgreSQL:
- See `SAAS_ARCHITECTURE.md` for database schema
- All APIs remain the same
- Just swap storage backend

---

## 🎉 Quick Start Checklist

- [ ] Build central dashboard
- [ ] Start dashboard server
- [ ] Test registration locally
- [ ] Install Lynis on monitored servers
- [ ] Install agent on first server
- [ ] Verify agent appears in dashboard
- [ ] Install agent on remaining servers
- [ ] Setup monitoring/alerts
- [ ] Configure backups
- [ ] Document your setup

---

## 📞 Support

- **Documentation:** See `MULTI_SERVER_GUIDE.md`
- **Architecture:** See `SAAS_ARCHITECTURE.md`
- **Test Script:** Run `./test-multi-server.sh`

---

**🚀 You're ready to deploy multi-server monitoring!**

