# 🛡️ Linux Hardening Dashboard

A professional, real-time security dashboard for Linux system hardening analysis built with Go and modern web technologies.

![Dashboard Preview](https://img.shields.io/badge/Status-Active-brightgreen)
![Go Version](https://img.shields.io/badge/Go-1.20+-blue)
![License](https://img.shields.io/badge/License-MIT-green)

## ✨ Features

### 🎯 **Core Functionality**
- **Real-time Security Analysis** - Live system security monitoring
- **Comprehensive Data Display** - All security metrics visible at once
- **Professional Black Theme** - Sleek cybersecurity aesthetic
- **Responsive Design** - Works on desktop, tablet, and mobile
- **Auto-refresh** - Updates every 60 seconds automatically

### 📊 **Security Sections**
- **🖥️ System Information** - OS, kernel, hardware details
- **🛡️ Security Status** - Hardening index, firewall, malware detection
- **🌐 Network Configuration** - IP addresses, DNS, interfaces
- **🔐 Authentication** - User accounts, password policies
- **📝 Logging & Auditing** - System logs and audit configuration
- **📦 Package Management** - Installed software and vulnerabilities
- **⚙️ Services & Daemons** - Running services status
- **📁 File System** - Directory permissions and binary analysis
- **🔒 Cryptography** - SSL certificates and encryption status
- **🧪 Test Results** - Security test execution details

### 🚀 **Advanced Features**
- **Smart Search** - Instant filtering across all security data
- **Export Functionality** - Download complete reports as JSON
- **One-Click Audits** - Run new security scans directly from dashboard
- **Status Indicators** - Real-time system health monitoring
- **Professional UI** - Enterprise-grade interface design

## 🏗️ Architecture

### Backend (Go)
- **Pure Go Implementation** - No external dependencies
- **Multi-location File Search** - Automatically finds Lynis reports
- **RESTful API** - Clean JSON endpoints
- **Embedded Templates** - Self-contained deployment

### Frontend
- **Modern JavaScript** - ES6+ features
- **CSS Grid & Flexbox** - Responsive layout system
- **Smooth Animations** - Professional UI transitions
- **Dark Theme** - Cybersecurity-focused design

## 🚀 Quick Start

### Prerequisites
- Go 1.20 or higher
- Lynis security auditing tool
- Modern web browser

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Pranavram22/UbuntuShield.git
   cd UbuntuShield
   ```

2. **Run the dashboard**
   ```bash
   go run main.go
   ```

3. **Access the dashboard**
   ```
   Open your browser to: http://localhost:5179
   ```

### First Time Setup

If you don't have a Lynis report yet:

1. **Install Lynis** (if not already installed)
   ```bash
   # Ubuntu/Debian
   sudo apt install lynis
   
   # macOS
   brew install lynis
   
   # CentOS/RHEL
   sudo yum install lynis
   ```

2. **Run a security audit**
   ```bash
   sudo lynis audit system
   ```

3. **Or use the dashboard's built-in audit button** 🚀

## 📡 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Main dashboard interface |
| `/report` | GET | JSON security data |
| `/run-audit` | POST | Trigger new security audit |

### API Response Example
```json
{
  "data": {
    "hardening_index": "78",
    "warnings": "12",
    "suggestions": "25",
    "tests_performed": "245",
    "os": "Linux",
    "hostname": "server-01",
    ...
  }
}
```

## 🎨 Screenshots

### Main Dashboard
- **Professional black theme** with neon green accents
- **Comprehensive stats cards** showing key security metrics
- **Organized data sections** for easy navigation

### Security Overview
- **Real-time hardening index** visualization
- **System health indicators** with trend analysis
- **Detailed test results** and recommendations

## 🔧 Configuration

### Default Settings
- **Port**: 5179
- **Auto-refresh**: 60 seconds
- **Report Locations**: Multiple paths automatically searched

### Customization
The dashboard automatically searches for Lynis reports in:
- `/Users/apple/lynis-report.dat`
- `/tmp/lynis-report.dat`
- `/usr/local/var/log/lynis-report.dat`
- `/var/log/lynis-report.dat`
- `./lynis-report.dat`

## 🛠️ Development

### Project Structure
```
UbuntuShield/
├── main.go                 # Main application server
├── templates/
│   └── dashboard.html      # Dashboard interface
├── README.md              # This file
└── go.mod                 # Go module file
```

### Building for Production
```bash
go build -o linux-hardening-dashboard main.go
./linux-hardening-dashboard
```

### Docker Deployment
```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dashboard main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dashboard .
EXPOSE 5179
CMD ["./dashboard"]
```

## 🔒 Security Features

### Data Protection
- **Local Processing** - All data stays on your system
- **No External Calls** - Completely self-contained
- **Secure Headers** - Proper HTTP security headers

### Audit Capabilities
- **Comprehensive Scanning** - 200+ security checks
- **Real-time Monitoring** - Live system status updates
- **Vulnerability Detection** - Package and configuration analysis

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'Add amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Development Guidelines
- Follow Go best practices
- Maintain responsive design
- Add tests for new features
- Update documentation

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Lynis** - Excellent security auditing tool
- **Go Community** - Amazing standard library
- **Open Source Security** - Tools and methodologies

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/Pranavram22/UbuntuShield/issues)
- **Documentation**: [Wiki](https://github.com/Pranavram22/UbuntuShield/wiki)
- **Discussions**: [GitHub Discussions](https://github.com/Pranavram22/UbuntuShield/discussions)

---

**Made with ❤️ for Linux Security**

*Secure your systems with professional-grade monitoring and analysis.*