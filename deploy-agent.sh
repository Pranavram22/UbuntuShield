#!/bin/bash
# UbuntuShield Agent Deployment Script
# Usage: ./deploy-agent.sh user@remote-server

set -e

if [ $# -eq 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🚀 UbuntuShield Agent Deployment Script"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Usage: $0 user@remote-server [central-server-ip]"
    echo ""
    echo "Examples:"
    echo "  $0 root@192.168.1.50"
    echo "  $0 ubuntu@192.168.1.50 192.168.1.100"
    echo ""
    exit 1
fi

TARGET=$1
CENTRAL_IP=${2:-$(hostname -I | awk '{print $1}')}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 UbuntuShield Agent Deployment"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Target Server:   $TARGET"
echo "Central Server:  http://$CENTRAL_IP:5179"
echo ""

# Check if agent directory exists
if [ ! -d "agent" ]; then
    echo "❌ Error: agent directory not found!"
    echo "   Please run this script from the UbuntuShield root directory"
    exit 1
fi

# Build agent
echo "📦 Building agent binary..."
cd agent
GOOS=linux GOARCH=amd64 go build -o ubuntushield-agent agent.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build agent"
    exit 1
fi
echo "✅ Agent built successfully"

# Copy to remote server
echo ""
echo "📤 Copying agent to remote server..."
scp ubuntushield-agent $TARGET:/tmp/ubuntushield-agent
if [ $? -ne 0 ]; then
    echo "❌ Failed to copy agent to remote server"
    exit 1
fi
echo "✅ Agent copied successfully"

# Setup on remote server
echo ""
echo "🔧 Setting up agent on remote server..."
ssh $TARGET "bash -s" << EOF
set -e

# Create directory
echo "  → Creating directory..."
sudo mkdir -p /opt/ubuntushield-agent

# Move binary
echo "  → Installing binary..."
sudo mv /tmp/ubuntushield-agent /opt/ubuntushield-agent/
sudo chmod +x /opt/ubuntushield-agent/ubuntushield-agent

# Create config
echo "  → Creating config..."
cat << CONFIGEOF | sudo tee /opt/ubuntushield-agent/config.json > /dev/null
{
    "server_url": "http://$CENTRAL_IP:5179",
    "api_key": "",
    "server_id": ""
}
CONFIGEOF

# Install Lynis
echo "  → Installing Lynis..."
if command -v apt >/dev/null 2>&1; then
    sudo apt update >/dev/null 2>&1
    sudo apt install -y lynis >/dev/null 2>&1
elif command -v yum >/dev/null 2>&1; then
    sudo yum install -y lynis >/dev/null 2>&1
else
    echo "  ⚠️  Could not install Lynis automatically. Please install manually."
fi

# Create systemd service
echo "  → Creating systemd service..."
cat << SERVICEEOF | sudo tee /etc/systemd/system/ubuntushield-agent.service > /dev/null
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
SERVICEEOF

# Enable and start service
echo "  → Enabling and starting service..."
sudo systemctl daemon-reload
sudo systemctl enable ubuntushield-agent >/dev/null 2>&1
sudo systemctl start ubuntushield-agent

# Check status
sleep 2
if sudo systemctl is-active --quiet ubuntushield-agent; then
    echo "  ✅ Agent service is running"
else
    echo "  ❌ Agent service failed to start"
    exit 1
fi

EOF

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Setup failed on remote server"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Deployment Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "The agent is now running on $TARGET"
echo ""
echo "🔍 To check status:"
echo "   ssh $TARGET \"sudo systemctl status ubuntushield-agent\""
echo ""
echo "📋 To view logs:"
echo "   ssh $TARGET \"sudo journalctl -u ubuntushield-agent -f\""
echo ""
echo "🌐 View in dashboard:"
echo "   http://$CENTRAL_IP:5179/"
echo ""
echo "The server should appear in your dashboard within 1-2 minutes!"
echo ""

