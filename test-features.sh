#!/bin/bash

# Test script for new Historical Tracking & Scheduler features
# Run this after starting the UbuntuShield dashboard

BASE_URL="http://localhost:5179"

echo "🧪 Testing UbuntuShield New Features"
echo "======================================"
echo ""

# Test 1: Check scheduler status
echo "1️⃣ Checking Scheduler Status..."
curl -s "$BASE_URL/scheduler/status" | jq '.'
echo ""
echo ""

# Test 2: Check storage stats
echo "2️⃣ Checking Storage Statistics..."
curl -s "$BASE_URL/history/stats" | jq '.'
echo ""
echo ""

# Test 3: Enable daily scheduled audits
echo "3️⃣ Enabling Daily Scheduled Audits..."
curl -s -X POST "$BASE_URL/scheduler/config" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "interval": "daily",
    "run_on_startup": false,
    "quiet_mode": true
  }' | jq '.'
echo ""
echo ""

# Test 4: Get 30-day trend (if data exists)
echo "4️⃣ Fetching 30-Day Trend Data..."
curl -s "$BASE_URL/history/trend?period=30d" | jq '.'
echo ""
echo ""

# Test 5: Get historical records
echo "5️⃣ Fetching Historical Records..."
curl -s "$BASE_URL/history/records" | jq '.count, .records[0:2]'
echo ""
echo ""

# Test 6: Compare with previous audit (if exists)
echo "6️⃣ Comparing with Previous Audit..."
curl -s "$BASE_URL/history/compare" | jq '.'
echo ""
echo ""

echo "✅ Feature testing complete!"
echo ""
echo "💡 Tips:"
echo "  - Run 'sudo lynis audit system' to generate audit data"
echo "  - Or use: curl -X POST $BASE_URL/run-audit"
echo "  - View dashboard at: $BASE_URL"
echo ""

