package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	VERSION     = "1.0.0"
	CONFIG_FILE = "/etc/ubuntushield/agent.conf"
)

// AgentConfig holds agent configuration
type AgentConfig struct {
	ServerID      string `json:"server_id"`
	APIKey        string `json:"api_key"`
	DashboardURL  string `json:"dashboard_url"`
	AuditInterval int    `json:"audit_interval"` // minutes
	Hostname      string `json:"hostname"`
}

// AgentMetrics holds system and audit metrics
type AgentMetrics struct {
	ServerID        string                 `json:"server_id"`
	Timestamp       time.Time              `json:"timestamp"`
	Hostname        string                 `json:"hostname"`
	OS              string                 `json:"os"`
	Arch            string                 `json:"arch"`
	AgentVersion    string                 `json:"agent_version"`
	HardeningIndex  string                 `json:"hardening_index"`
	Warnings        string                 `json:"warnings"`
	TestsPerformed  string                 `json:"tests_performed"`
	ComplianceScore map[string]interface{} `json:"compliance_score"`
	RawData         map[string]string      `json:"raw_data"`
}

// RegistrationRequest for initial agent registration
type RegistrationRequest struct {
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	AgentVersion string `json:"agent_version"`
}

// RegistrationResponse from dashboard
type RegistrationResponse struct {
	Success  bool   `json:"success"`
	ServerID string `json:"server_id"`
	APIKey   string `json:"api_key"`
	Message  string `json:"message"`
}

// HeartbeatRequest for keepalive
type HeartbeatRequest struct {
	ServerID     string    `json:"server_id"`
	Timestamp    time.Time `json:"timestamp"`
	Status       string    `json:"status"`
	AgentVersion string    `json:"agent_version"`
}

// Agent represents the monitoring agent
type Agent struct {
	config *AgentConfig
	client *http.Client
}

// NewAgent creates a new agent instance
func NewAgent() *Agent {
	return &Agent{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LoadConfig loads agent configuration from file
func (a *Agent) LoadConfig() error {
	// Try to load existing config
	data, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		// Config doesn't exist, will need to register
		return err
	}

	config := &AgentConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	a.config = config
	log.Printf("✅ Config loaded: Server ID: %s\n", config.ServerID)
	return nil
}

// SaveConfig saves agent configuration to file
func (a *Agent) SaveConfig() error {
	// Create directory if not exists
	os.MkdirAll("/etc/ubuntushield", 0755)

	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(CONFIG_FILE, data, 0600)
}

// Register registers the agent with the central dashboard
func (a *Agent) Register(dashboardURL string) error {
	log.Println("📝 Registering agent with dashboard...")

	hostname, _ := os.Hostname()

	req := RegistrationRequest{
		Hostname:     hostname,
		IPAddress:    getOutboundIP(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		AgentVersion: VERSION,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := a.client.Post(
		dashboardURL+"/api/agents/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to dashboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(bodyBytes))
	}

	var regResp RegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return err
	}

	if !regResp.Success {
		return fmt.Errorf("registration failed: %s", regResp.Message)
	}

	// Save configuration
	a.config = &AgentConfig{
		ServerID:      regResp.ServerID,
		APIKey:        regResp.APIKey,
		DashboardURL:  dashboardURL,
		AuditInterval: 60, // 60 minutes default
		Hostname:      hostname,
	}

	if err := a.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Printf("✅ Registration successful!")
	log.Printf("   Server ID: %s\n", regResp.ServerID)
	log.Printf("   API Key: %s...\n", regResp.APIKey[:20])

	return nil
}

// SendHeartbeat sends keepalive to dashboard
func (a *Agent) SendHeartbeat() error {
	req := HeartbeatRequest{
		ServerID:     a.config.ServerID,
		Timestamp:    time.Now(),
		Status:       "active",
		AgentVersion: VERSION,
	}

	return a.sendRequest("/api/agents/heartbeat", req)
}

// RunAudit executes Lynis and sends results to dashboard
func (a *Agent) RunAudit() error {
	log.Println("🔍 Starting security audit...")

	// Check if Lynis is installed
	_, err := exec.LookPath("lynis")
	if err != nil {
		return fmt.Errorf("Lynis not found. Please install Lynis first")
	}

	// Run Lynis audit
	cmd := exec.Command("sudo", "lynis", "audit", "system", "--quick", "--quiet")
	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("⚠️ Lynis audit completed with warnings: %v\n", err)
	} else {
		log.Println("✅ Lynis audit completed")
	}

	// Parse Lynis report
	data, err := parseLynisReport()
	if err != nil {
		return fmt.Errorf("failed to parse Lynis report: %w", err)
	}

	// Prepare metrics
	metrics := AgentMetrics{
		ServerID:       a.config.ServerID,
		Timestamp:      time.Now(),
		Hostname:       a.config.Hostname,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		AgentVersion:   VERSION,
		HardeningIndex: data["hardening_index"],
		Warnings:       data["warnings"],
		TestsPerformed: data["lynis_tests_done"],
		RawData:        data,
	}

	// Send metrics to dashboard
	log.Println("📤 Sending audit results to dashboard...")
	if err := a.sendRequest("/api/metrics", metrics); err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	log.Println("✅ Audit results sent successfully")
	log.Printf("   Hardening Index: %s%%\n", data["hardening_index"])
	log.Printf("   Warnings: %s\n", data["warnings"])

	return nil
}

// sendRequest sends authenticated request to dashboard
func (a *Agent) sendRequest(endpoint string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		a.config.DashboardURL+endpoint,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s", string(bodyBytes))
	}

	return nil
}

// Run starts the agent main loop
func (a *Agent) Run() {
	log.Printf("🚀 UbuntuShield Agent v%s starting...\n", VERSION)

	// Send initial heartbeat
	if err := a.SendHeartbeat(); err != nil {
		log.Printf("⚠️ Failed to send heartbeat: %v\n", err)
	}

	// Run initial audit
	if err := a.RunAudit(); err != nil {
		log.Printf("❌ Initial audit failed: %v\n", err)
	}

	// Start periodic tasks
	heartbeatTicker := time.NewTicker(5 * time.Minute)
	auditTicker := time.NewTicker(time.Duration(a.config.AuditInterval) * time.Minute)

	defer heartbeatTicker.Stop()
	defer auditTicker.Stop()

	log.Printf("⏰ Heartbeat interval: 5 minutes\n")
	log.Printf("⏰ Audit interval: %d minutes\n", a.config.AuditInterval)
	log.Println("✅ Agent is running. Press Ctrl+C to stop.")

	for {
		select {
		case <-heartbeatTicker.C:
			if err := a.SendHeartbeat(); err != nil {
				log.Printf("⚠️ Heartbeat failed: %v\n", err)
			} else {
				log.Println("💓 Heartbeat sent")
			}

		case <-auditTicker.C:
			if err := a.RunAudit(); err != nil {
				log.Printf("❌ Audit failed: %v\n", err)
			}
		}
	}
}

// Helper functions

func parseLynisReport() (map[string]string, error) {
	reportPaths := []string{
		"/tmp/lynis-report.dat",
		"/var/log/lynis-report.dat",
		"/usr/local/var/log/lynis-report.dat",
	}

	for _, path := range reportPaths {
		data, err := parseLynisFile(path)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}

	return nil, fmt.Errorf("no Lynis report found")
}

func parseLynisFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			data[key] = value
		}
	}

	return data, scanner.Err()
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
	log.SetFlags(log.Ltime)

	if len(os.Args) < 2 {
		fmt.Println("UbuntuShield Agent v" + VERSION)
		fmt.Println("\nUsage:")
		fmt.Println("  agent register <dashboard-url>  - Register with central dashboard")
		fmt.Println("  agent start                      - Start the agent")
		fmt.Println("  agent audit                      - Run a single audit")
		fmt.Println("  agent status                     - Show agent status")
		fmt.Println("\nExample:")
		fmt.Println("  agent register https://dashboard.example.com")
		fmt.Println("  agent start")
		os.Exit(1)
	}

	agent := NewAgent()
	command := os.Args[1]

	switch command {
	case "register":
		if len(os.Args) < 3 {
			log.Fatal("❌ Dashboard URL required")
		}
		dashboardURL := os.Args[2]

		if err := agent.Register(dashboardURL); err != nil {
			log.Fatalf("❌ Registration failed: %v\n", err)
		}

	case "start":
		// Load config
		if err := agent.LoadConfig(); err != nil {
			log.Fatal("❌ Config not found. Please run 'agent register' first")
		}

		// Run agent
		agent.Run()

	case "audit":
		// Load config
		if err := agent.LoadConfig(); err != nil {
			log.Fatal("❌ Config not found. Please run 'agent register' first")
		}

		// Run single audit
		if err := agent.RunAudit(); err != nil {
			log.Fatalf("❌ Audit failed: %v\n", err)
		}

	case "status":
		// Load config
		if err := agent.LoadConfig(); err != nil {
			log.Fatal("❌ Agent not configured")
		}

		fmt.Printf("UbuntuShield Agent v%s\n", VERSION)
		fmt.Printf("Server ID: %s\n", agent.config.ServerID)
		fmt.Printf("Dashboard: %s\n", agent.config.DashboardURL)
		fmt.Printf("Hostname: %s\n", agent.config.Hostname)
		fmt.Printf("Audit Interval: %d minutes\n", agent.config.AuditInterval)

	default:
		log.Fatal("❌ Unknown command: " + command)
	}
}

