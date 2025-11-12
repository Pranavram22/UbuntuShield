# 📈 New Features: Historical Tracking & Automated Audits

## Overview

Your UbuntuShield now includes two powerful new features:
1. **Historical Tracking & Trend Analysis** - Monitor security improvements over time
2. **Automated Scheduled Audits** - Set-it-and-forget-it security monitoring

---

## 📊 Historical Tracking & Trend Analysis

### What It Does
- **Automatically saves** every audit result
- **Tracks security scores** over time
- **Compares** current vs previous audits
- **Generates trend charts** for visualization
- **Efficient storage** with compression

### Data Usage & Storage

#### Storage Efficiency
```
Single Audit Record: ~5-7 KB (compressed)
Daily Audits for 1 Year: ~2.5 MB total
Monthly Audits for 5 Years: ~500 KB total
```

#### Automatic Management
- ✅ **Auto-compression** after 30 days
- ✅ **Auto-cleanup** after 365 days
- ✅ **Smart delta storage** (only stores changes)
- ✅ **Local storage only** (no bandwidth usage)

#### Storage Locations
```
./history/                    # Main storage directory
├── audit_2025-01-15_10-30-00.json      # Recent (uncompressed)
├── audit_2025-01-14_10-30-00.json      # Recent (uncompressed)
└── audit_2024-12-15_10-30-00.json.gz   # Older (compressed)
```

### API Endpoints

#### Get Trend Data
```bash
GET /history/trend?period=30d

# Periods: 7d, 30d, 90d

Response:
{
  "period": "30d",
  "security_score_trend": [
    {"timestamp": "2025-01-01T10:00:00Z", "value": 75},
    {"timestamp": "2025-01-02T10:00:00Z", "value": 78},
    ...
  ],
  "warnings_trend": [...],
  "tests_trend": [...]
}
```

#### Get Historical Records
```bash
GET /history/records?since=2025-01-01T00:00:00Z

Response:
{
  "records": [...],
  "count": 15,
  "since": "2025-01-01T00:00:00Z"
}
```

#### Compare with Previous Audit
```bash
GET /history/compare

Response:
{
  "success": true,
  "comparison": {
    "score_change": 3.5,
    "warnings_change": -2,
    "improved": true,
    "previous_date": "2025-01-14T10:00:00Z",
    "days_since": 1.2
  }
}
```

#### Storage Statistics
```bash
GET /history/stats

Response:
{
  "total_records": 90,
  "compressed_count": 60,
  "total_size_mb": 2.3,
  "avg_record_size": 25600,
  "compression_ratio": 66.7
}
```

---

## ⏰ Automated Scheduled Audits

### What It Does
- **Automatically runs** Lynis audits on schedule
- **No manual intervention** required
- **Runs in background** without interrupting system
- **Automatically saves** results to history
- **Configurable schedules** (hourly, daily, weekly, monthly)

### How It Works

#### Automatic Execution Flow
```
1. Scheduler starts when application launches
2. Timer triggers at configured interval
3. Checks if Lynis is installed
4. Runs: sudo lynis audit system --quick --quiet
5. Lynis generates .dat file automatically to /tmp/lynis-report.dat
6. Application reads the .dat file
7. Parses and analyzes results
8. Saves to history
9. Ready for next scheduled run
```

#### No Bandwidth Used
- ✅ Everything runs **locally** on your machine
- ✅ No external API calls
- ✅ No data sent over network
- ✅ All files stored on disk

### Configuration

#### Via API

**Get Status:**
```bash
GET /scheduler/status

Response:
{
  "enabled": true,
  "running": true,
  "interval": "24h0m0s",
  "last_run": "2025-01-15T10:30:00Z",
  "next_run": "2025-01-16T10:30:00Z",
  "run_on_startup": false,
  "quiet_mode": true
}
```

**Update Configuration:**
```bash
POST /scheduler/config
Content-Type: application/json

{
  "enabled": true,
  "interval": "daily",
  "run_on_startup": false,
  "quiet_mode": true
}

# Intervals: "hourly", "daily", "weekly", "monthly"
```

#### Preset Schedules

**Hourly** (for testing/development)
```json
{
  "enabled": true,
  "interval": "hourly",
  "quiet_mode": true
}
```

**Daily** (recommended for production)
```json
{
  "enabled": true,
  "interval": "daily",
  "quiet_mode": true
}
```

**Weekly** (for stable systems)
```json
{
  "enabled": true,
  "interval": "weekly",
  "quiet_mode": true
}
```

**Monthly** (for minimal overhead)
```json
{
  "enabled": true,
  "interval": "monthly",
  "quiet_mode": true
}
```

---

## 🚀 Getting Started

### 1. Run the Application
```bash
go run main.go
```

You'll see:
```
💾 History manager initialized
⏰ Audit scheduler initialized
⏰ Scheduler is disabled. Enable it in settings to run automated audits.
🚀 Linux Hardening Dashboard starting on http://localhost:5179
```

### 2. Enable Scheduled Audits (Optional)
```bash
# Enable daily automated audits
curl -X POST http://localhost:5179/scheduler/config \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "interval": "daily",
    "run_on_startup": false,
    "quiet_mode": true
  }'
```

### 3. Run Manual Audit
```bash
# Or run audit manually to start tracking
curl -X POST http://localhost:5179/run-audit
```

### 4. View History
```bash
# Get last 30 days trend
curl http://localhost:5179/history/trend?period=30d

# View storage stats
curl http://localhost:5179/history/stats
```

---

## 📊 Example Usage Scenarios

### Scenario 1: Daily Security Monitoring
```bash
# Setup
curl -X POST http://localhost:5179/scheduler/config \
  -d '{"enabled":true,"interval":"daily","quiet_mode":true}'

# Result:
# - Audit runs every day at the same time
# - Results saved automatically
# - Trends available on dashboard
# - Storage: ~2.5 MB per year
```

### Scenario 2: Compare Before/After System Changes
```bash
# Before making changes
curl -X POST http://localhost:5179/run-audit

# Make your system changes...

# After changes
curl -X POST http://localhost:5179/run-audit

# Compare
curl http://localhost:5179/history/compare
```

### Scenario 3: Weekly Compliance Reports
```bash
# Enable weekly audits
curl -X POST http://localhost:5179/scheduler/config \
  -d '{"enabled":true,"interval":"weekly"}'

# Generate trend report
curl http://localhost:5179/history/trend?period=90d > report.json
```

---

## 🔍 Technical Details

### How Lynis .dat File is Generated

**Automatic Generation:**
When you run `sudo lynis audit system`, Lynis automatically:
1. Performs all security checks
2. Generates `lynis-report.dat` in `/tmp/`
3. Also copies to `/var/log/lynis-report.dat` (if permissions allow)

**Your Application:**
- Searches multiple locations for the `.dat` file
- Parses the file (simple key=value format)
- No need to manually move or copy files

### Scheduler Implementation

**Technology:**
- Go's native `time.Ticker` for scheduling
- Goroutines for background execution
- No external cron dependencies

**Process:**
```go
// Pseudo-code
scheduler.Start() {
  ticker := time.NewTicker(interval)
  for {
    <-ticker.C
    runAudit()
    saveToHistory()
  }
}
```

### Storage Compression

**Compression Algorithm:**
- Uses Go's `compress/gzip` (standard library)
- Compression ratio: ~70-80%
- Automatic after 30 days
- Transparent decompression when reading

---

## 🎯 Benefits

### 1. Zero Bandwidth Usage
- ✅ All processing is local
- ✅ No external API calls
- ✅ No data transmission
- ✅ Works offline

### 2. Minimal Storage
- ✅ ~2.5 MB per year of daily audits
- ✅ Automatic compression
- ✅ Automatic cleanup
- ✅ Smart delta storage

### 3. Set-and-Forget
- ✅ Automatic audit execution
- ✅ No manual intervention
- ✅ Background processing
- ✅ Logs all activities

### 4. Trend Analysis
- ✅ Track security improvements
- ✅ Identify degradations quickly
- ✅ Historical compliance data
- ✅ Visual charts (coming soon)

---

## 🔒 Security Considerations

### Permissions
- Lynis requires `sudo` to run comprehensive audits
- History files stored with user permissions (0755)
- No sensitive data exposed via API

### Privacy
- All data stays on your local machine
- No external connections
- No telemetry or tracking
- You control all data

### Reliability
- Graceful error handling
- Failed audits don't crash scheduler
- Corrupted history files are skipped
- Automatic retry on next schedule

---

## 📝 Troubleshooting

### Issue: Scheduler Not Running
```bash
# Check status
curl http://localhost:5179/scheduler/status

# Enable if needed
curl -X POST http://localhost:5179/scheduler/config \
  -d '{"enabled":true,"interval":"daily"}'
```

### Issue: No History Data
```bash
# Check storage stats
curl http://localhost:5179/history/stats

# Run manual audit to create first record
curl -X POST http://localhost:5179/run-audit
```

### Issue: Lynis Not Found
```bash
# Install Lynis
# Ubuntu/Debian:
sudo apt install lynis

# macOS:
brew install lynis

# Verify installation
which lynis
```

### Issue: Permission Denied
```bash
# Ensure user can run sudo without password for lynis
# Add to /etc/sudoers (use visudo):
your_username ALL=(ALL) NOPASSWD: /usr/bin/lynis
```

---

## 🎨 Future Enhancements (Coming Soon)

- [ ] Web UI for trend visualization (charts)
- [ ] Email notifications on score drops
- [ ] Export history as PDF reports
- [ ] Custom retention policies
- [ ] Webhook integration
- [ ] Compare multiple audits side-by-side

---

## 📞 Questions?

Check the main README.md for general information and support channels.

**Happy monitoring! 🛡️**

