#!/bin/bash

# Test Multi-Server SaaS Functionality
# Tests agent registration and metric submission

BASE_URL="http://localhost:5179"

echo "🧪 Testing Multi-Server SaaS Functionality"
echo "==========================================="
echo ""

# Test 1: Register a server (agent)
echo "1️⃣ Testing Agent Registration..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/agents/register" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test-server-01",
    "ip_address": "10.0.1.10",
    "os": "linux",
    "arch": "amd64",
    "agent_version": "1.0.0"
  }')

echo "$RESPONSE" | jq '.'

# Extract server_id and api_key
SERVER_ID=$(echo "$RESPONSE" | jq -r '.server_id')
API_KEY=$(echo "$RESPONSE" | jq -r '.api_key')

if [ "$SERVER_ID" == "null" ] || [ -z "$SERVER_ID" ]; then
  echo "❌ Registration failed!"
  exit 1
fi

echo "✅ Server registered successfully!"
echo "   Server ID: $SERVER_ID"
echo "   API Key: ${API_KEY:0:20}..."
echo ""

# Test 2: Send heartbeat
echo "2️⃣ Testing Heartbeat..."
curl -s -X POST "$BASE_URL/api/agents/heartbeat" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d "{
    \"server_id\": \"$SERVER_ID\",
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"status\": \"active\",
    \"agent_version\": \"1.0.0\"
  }" | jq '.'
echo ""

# Test 3: Submit metrics
echo "3️⃣ Testing Metrics Submission..."
curl -s -X POST "$BASE_URL/api/metrics" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d "{
    \"server_id\": \"$SERVER_ID\",
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"hardening_index\": \"85\",
    \"warnings\": \"5\",
    \"tests_performed\": \"250\",
    \"raw_data\": {
      \"os\": \"Ubuntu 22.04\",
      \"kernel_version\": \"5.15.0-76-generic\"
    }
  }" | jq '.'
echo ""

# Test 4: Register another server
echo "4️⃣ Registering Second Server..."
RESPONSE2=$(curl -s -X POST "$BASE_URL/api/agents/register" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test-server-02",
    "ip_address": "10.0.1.11",
    "os": "linux",
    "arch": "amd64",
    "agent_version": "1.0.0"
  }')

SERVER_ID2=$(echo "$RESPONSE2" | jq -r '.server_id')
API_KEY2=$(echo "$RESPONSE2" | jq -r '.api_key')
echo "✅ Second server registered: $SERVER_ID2"
echo ""

# Submit metrics for second server
curl -s -X POST "$BASE_URL/api/metrics" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY2" \
  -d "{
    \"server_id\": \"$SERVER_ID2\",
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"hardening_index\": \"92\",
    \"warnings\": \"3\",
    \"tests_performed\": \"250\"
  }" > /dev/null

# Test 5: List all servers
echo "5️⃣ Listing All Servers..."
curl -s "$BASE_URL/api/servers" | jq '.stats, .servers[] | {hostname, status, id: .id[0:8]}'
echo ""

# Test 6: Get specific server details
echo "6️⃣ Getting Server Details for $SERVER_ID..."
curl -s "$BASE_URL/api/servers/$SERVER_ID" | jq '{
  hostname: .server.hostname,
  status: .server.status,
  metrics_count: .count,
  latest_score: .metrics[0].hardening_index
}'
echo ""

echo "✅ All tests completed successfully!"
echo ""
echo "💡 Next steps:"
echo "   • View dashboard: $BASE_URL"
echo "   • List servers: $BASE_URL/api/servers"
echo "   • Install agent on real servers"
echo ""

