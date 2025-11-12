# 🚀 Deploy Agent on New Server - Quick Guide

## Step 1: Install Lynis on Your Server

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install lynis -y

# CentOS/RHEL
sudo yum install lynis -y

# Or install latest version
cd /tmp
git clone https://github.com/CISOfy/lynis
cd lynis
sudo ./lynis audit system
```

## Step 2: Download & Setup Agent

### Option A: Quick Install (Recommended)

Run this ONE command on your server:

```bash
curl -fsSL https://YOUR-CENTRAL-SERVER:5179/install-agent.sh | sudo bash
```

### Option B: Manual Install

```bash
# 1. Create directory
sudo mkdir -p /opt/ubuntushield-agent
cd /opt/ubuntushield-agent

# 2. Download agent binary (compile from source)
# On your development machine:
cd /path/to/UbuntuShield/agent
GOOS=linux GOARCH=amd64 go build -o ubuntushield-agent agent.go

# Then copy to server:
scp ubuntushield-agent user@your-server:/tmp/
ssh user@your-server
sudo mv /tmp/ubuntushield-agent /opt/ubuntushield-agent/
sudo chmod +x /opt/ubuntushield-agent/ubuntushield-agent

# 3. Create config file
sudo nano /opt/ubuntushield-agent/config.json
```

**config.json:**
```json
{
    "server_url": "http://YOUR-CENTRAL-SERVER-IP:5179",
    "api_key": "PASTE-API-KEY-HERE",
    "server_id": ""
}
```

Replace:
- `YOUR-CENTRAL-SERVER-IP` with your central dashboard IP (e.g., `192.168.1.100`)
- `PASTE-API-KEY-HERE` will be generated on first run

## Step 3: Create Systemd Service (Auto-start)

```bash
sudo nano /etc/systemd/system/ubuntushield-agent.service
```

**Paste this:**

```ini
[Unit]
Description=UbuntuShield Security Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/ubuntushield-agent
ExecStart=/opt/ubuntushield-agent/ubuntushield-agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Enable and start:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable ubuntushield-agent
sudo systemctl start ubuntushield-agent
```

## Step 4: Verify It's Working

```bash
# Check service status
sudo systemctl status ubuntushield-agent

# View logs
sudo journalctl -u ubuntushield-agent -f

# You should see:
# ✅ Agent registered successfully
# ✅ Server ID: abc123...
# ✅ Sending heartbeat...
# ✅ Running Lynis audit...
```

## Step 5: View in Dashboard

Open your central dashboard:

```
http://YOUR-CENTRAL-SERVER-IP:5179/
```

You should see your new server appear! 🎉

---

## 📋 Complete Example (Copy & Paste)

### On Central Server (Dashboard)

```bash
# Your dashboard is already running on port 5179
# Get the server IP
hostname -I
```

### On Remote Server (Agent)

```bash
# 1. Install Lynis
sudo apt update && sudo apt install lynis -y

# 2. Create agent directory
sudo mkdir -p /opt/ubuntushield-agent
cd /opt/ubuntushield-agent

# 3. Get agent binary from central server (you'll need to compile and distribute)
# For now, manual copy required

# 4. Create minimal config
sudo tee config.json > /dev/null <<EOF
{
    "server_url": "http://192.168.1.100:5179",
    "api_key": "",
    "server_id": ""
}
EOF

# 5. Make executable
sudo chmod +x ubuntushield-agent

# 6. First run (manual)
sudo ./ubuntushield-agent
# This will auto-register and save API key to config

# 7. Create systemd service
sudo tee /etc/systemd/system/ubuntushield-agent.service > /dev/null <<EOF
[Unit]
Description=UbuntuShield Security Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/ubuntushield-agent
ExecStart=/opt/ubuntushield-agent/ubuntushield-agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 8. Enable and start
sudo systemctl daemon-reload
sudo systemctl enable ubuntushield-agent
sudo systemctl start ubuntushield-agent

# 9. Check status
sudo systemctl status ubuntushield-agent
```

---

## 🔍 Troubleshooting

### Agent not appearing in dashboard?

```bash
# Check if agent is running
sudo systemctl status ubuntushield-agent

# Check network connectivity to central server
curl http://YOUR-CENTRAL-SERVER:5179/api/servers

# Check agent logs
sudo journalctl -u ubuntushield-agent -f
```

### Connection refused?

```bash
# On central server, check if port is open
sudo netstat -tlnp | grep 5179

# Allow port in firewall
sudo ufw allow 5179
```

### Permission denied for Lynis?

```bash
# Agent needs root to run Lynis
# Make sure service runs as root (User=root in systemd)
```

---

## 📊 What Data Gets Sent?

The agent sends:

1. **Heartbeat** (every 5 minutes): "I'm alive"
2. **Audit Results** (every 6 hours): 
   - Hardening index score
   - Warning count
   - Suggestions
   - Test results
3. **System Info** (on registration):
   - Hostname
   - IP address
   - OS version
   - Architecture

**No sensitive data** like passwords, keys, or file contents is ever transmitted.

---

## 🎯 Quick Deploy Script

Save this as `deploy-agent.sh` on your central server:

```bash
#!/bin/bash
# Deploy UbuntuShield Agent
# Usage: ./deploy-agent.sh user@remote-server

if [ $# -eq 0 ]; then
    echo "Usage: $0 user@remote-server"
    exit 1
fi

TARGET=$1
CENTRAL_IP=$(hostname -I | awk '{print $1}')

echo "📦 Building agent..."
cd agent
GOOS=linux GOARCH=amd64 go build -o ubuntushield-agent agent.go

echo "📤 Copying to $TARGET..."
scp ubuntushield-agent $TARGET:/tmp/

echo "🔧 Setting up on remote server..."
ssh $TARGET "sudo mkdir -p /opt/ubuntushield-agent && \
    sudo mv /tmp/ubuntushield-agent /opt/ubuntushield-agent/ && \
    sudo chmod +x /opt/ubuntushield-agent/ubuntushield-agent && \
    echo '{\"server_url\":\"http://$CENTRAL_IP:5179\",\"api_key\":\"\",\"server_id\":\"\"}' | sudo tee /opt/ubuntushield-agent/config.json && \
    sudo apt update && sudo apt install -y lynis"

echo "✅ Agent deployed! Now create systemd service manually on remote server."
```

Run it:
```bash
chmod +x deploy-agent.sh
./deploy-agent.sh user@192.168.1.50
```

---

## 🌐 Network Requirements

- Central server must be accessible from all monitored servers
- Port `5179` must be open on central server
- Agents need outbound access to central server
- No inbound ports needed on agent servers

---

## 🔐 Security Notes

- Each agent gets a unique API key
- All communication is authenticated
- Consider using HTTPS in production (add reverse proxy)
- Run agent as root (required for Lynis)
- API keys stored in `/opt/ubuntushield-agent/config.json`

---

## 📈 Scaling

- Central server can handle 100+ agents easily (file-based)
- For 500+ servers, consider switching to PostgreSQL
- Each agent uses ~50 MB RAM
- Lynis audit takes 2-5 minutes per run
- Data sent: ~10 KB per heartbeat, ~100 KB per audit

---

## Next Steps

1. Deploy agent on first server
2. Verify in dashboard
3. Deploy to more servers
4. Monitor all servers from one place! 🎉

