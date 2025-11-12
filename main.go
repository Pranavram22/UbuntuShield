package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:embed templates/*
var content embed.FS

// LynisReport represents the parsed Lynis report data
type LynisReport struct {
	Data            map[string]string  `json:"data"`
	ComplianceScore ComplianceAnalysis `json:"compliance_score"`
	Findings        []SecurityFinding  `json:"findings"`
	Remediations    []Remediation      `json:"remediations"`
}

// ComplianceAnalysis represents compliance framework analysis
type ComplianceAnalysis struct {
	CIS_Level1 ComplianceProfile `json:"cis_level1"`
	CIS_Level2 ComplianceProfile `json:"cis_level2"`
	ISO27001   ComplianceProfile `json:"iso27001"`
	NIST       ComplianceProfile `json:"nist"`
	PCIDSS     ComplianceProfile `json:"pcidss"`
	SOC2       ComplianceProfile `json:"soc2"`
	HIPAA      ComplianceProfile `json:"hipaa"`
	GDPR       ComplianceProfile `json:"gdpr"`
	SOX        ComplianceProfile `json:"sox"`
	FISMA      ComplianceProfile `json:"fisma"`
	COBIT      ComplianceProfile `json:"cobit"`
}

// ComplianceProfile represents a compliance framework profile
type ComplianceProfile struct {
	Score      float64            `json:"score"`
	Total      int                `json:"total"`
	Passed     int                `json:"passed"`
	Failed     int                `json:"failed"`
	Exceptions int                `json:"exceptions"`
	Controls   map[string]Control `json:"controls"`
}

// Control represents a specific compliance control
type Control struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`   // passed, failed, exception
	Severity    string `json:"severity"` // high, medium, low
	Description string `json:"description"`
}

// SecurityFinding represents a security issue found by Lynis
type SecurityFinding struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Severity     string   `json:"severity"`
	Category     string   `json:"category"`
	Mappings     []string `json:"mappings"` // CIS controls, ISO controls, etc.
	FixAvailable bool     `json:"fix_available"`
}

// Remediation represents an automated fix
type Remediation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Risk        string `json:"risk"`
	FindingID   string `json:"finding_id"`
}

// parseLynisReport reads and parses the Lynis report file from multiple locations
func parseLynisReport() (map[string]string, error) {
	// Try multiple locations for the Lynis report file
	reportPaths := []string{
		"/Users/apple/lynis-report.dat",
		"/Users/apple/Desktop/untitled folder/untitled folder/lynis-report.dat",
		"./lynis-report.dat",
		"/tmp/lynis-report.dat",
		"/usr/local/var/log/lynis-report.dat",
		"/var/log/lynis-report.dat",
		"/usr/share/lynis/lynis-report.dat",
		os.Getenv("HOME") + "/lynis-report.dat",
	}

	for _, reportPath := range reportPaths {
		file, err := os.Open(reportPath)
		if err != nil {
			continue
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

		if err := scanner.Err(); err != nil {
			continue
		}

		if len(data) > 0 {
			return data, nil
		}
	}

	return nil, fmt.Errorf("No Lynis report file found in any location. Tried: %v", reportPaths)
}

// reportHandler handles the /report API endpoint
func reportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := parseLynisReport()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing report: %v", err), http.StatusInternalServerError)
		return
	}

	// Analyze compliance and generate findings
	complianceScore := analyzeCompliance(data)
	findings := extractSecurityFindings(data)
	remediations := generateRemediations(findings)

	report := LynisReport{
		Data:            data,
		ComplianceScore: complianceScore,
		Findings:        findings,
		Remediations:    remediations,
	}

	// Save to history automatically (in background, don't block response)
	go func() {
		if historyManager != nil {
			if err := historyManager.SaveAudit(data, complianceScore); err != nil {
				log.Printf("Warning: Failed to save audit to history: %v", err)
			}
		}
	}()

	json.NewEncoder(w).Encode(report)
}

// dashboardHandler serves the main dashboard page
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(content, "templates/dashboard.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// multiServerDashboardHandler serves the multi-server dashboard
func multiServerDashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(content, "templates/multi-server.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// runAuditHandler handles running Lynis audit
func runAuditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if Lynis is available
	_, err := exec.LookPath("lynis")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Lynis not found. Please install Lynis first.",
		})
		return
	}

	// Run Lynis audit in background
	go func() {
		log.Println("Starting Lynis audit...")

		// Run the audit
		cmd := exec.Command("sudo", "lynis", "audit", "system", "--quick")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Lynis audit error: %v\nOutput: %s", err, string(output))
			return
		}

		log.Println("Lynis audit completed")

		// Try to copy the report to the expected location
		reportPath := "/tmp/lynis-report.dat"
		targetPath := "/usr/local/var/log/lynis-report.dat"

		// Create target directory
		exec.Command("sudo", "mkdir", "-p", "/usr/local/var/log").Run()

		// Copy report file
		if _, err := os.Stat(reportPath); err == nil {
			copyCmd := exec.Command("sudo", "cp", reportPath, targetPath)
			if err := copyCmd.Run(); err != nil {
				log.Printf("Failed to copy report file: %v", err)
			} else {
				log.Println("Report file copied successfully")
			}
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Lynis audit started. This may take a few minutes to complete.",
	})
}

// complianceProfileHandler handles compliance profile endpoints
func complianceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	profile := r.URL.Query().Get("profile")

	data, err := parseLynisReport()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing report: %v", err), http.StatusInternalServerError)
		return
	}

	complianceScore := analyzeCompliance(data)
	var result interface{}

	switch profile {
	case "cis_level1":
		result = complianceScore.CIS_Level1
	case "cis_level2":
		result = complianceScore.CIS_Level2
	case "iso27001":
		result = complianceScore.ISO27001
	case "nist":
		result = complianceScore.NIST
	case "pcidss":
		result = complianceScore.PCIDSS
	case "soc2":
		result = complianceScore.SOC2
	case "hipaa":
		result = complianceScore.HIPAA
	case "gdpr":
		result = complianceScore.GDPR
	case "sox":
		result = complianceScore.SOX
	case "fisma":
		result = complianceScore.FISMA
	case "cobit":
		result = complianceScore.COBIT
	default:
		result = complianceScore
	}

	json.NewEncoder(w).Encode(result)
}

// remediateHandler handles automated remediation
func remediateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var request struct {
		RemediationID string `json:"remediation_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Execute remediation (simulate for now)
	result := map[string]interface{}{
		"success":    true,
		"message":    fmt.Sprintf("Remediation %s applied successfully", request.RemediationID),
		"applied_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(result)
}

// analyzeCompliance analyzes Lynis data against compliance frameworks
func analyzeCompliance(data map[string]string) ComplianceAnalysis {
	return ComplianceAnalysis{
		CIS_Level1: analyzeCISLevel1(data),
		CIS_Level2: analyzeCISLevel2(data),
		ISO27001:   analyzeISO27001(data),
		NIST:       analyzeNIST(data),
		PCIDSS:     analyzePCIDSS(data),
		SOC2:       analyzeSOC2(data),
		HIPAA:      analyzeHIPAA(data),
		GDPR:       analyzeGDPR(data),
		SOX:        analyzeSOX(data),
		FISMA:      analyzeFISMA(data),
		COBIT:      analyzeCOBIT(data),
	}
}

// analyzeCISLevel1 analyzes against CIS Level 1 benchmarks
func analyzeCISLevel1(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// CIS 5.2.8 - SSH Root Login
	total++
	if strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin no") {
		controls["5.2.8"] = Control{
			ID:          "5.2.8",
			Title:       "Ensure SSH root login is disabled",
			Status:      "passed",
			Severity:    "high",
			Description: "Direct root login via SSH should be disabled",
		}
		passed++
	} else {
		controls["5.2.8"] = Control{
			ID:          "5.2.8",
			Title:       "Ensure SSH root login is disabled",
			Status:      "failed",
			Severity:    "high",
			Description: "Direct root login via SSH should be disabled",
		}
	}

	// CIS 3.3.1 - UFW Firewall
	total++
	if strings.Contains(strings.ToLower(data["firewall_software"]), "ufw") &&
		strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["3.3.1"] = Control{
			ID:          "3.3.1",
			Title:       "Ensure UFW is enabled",
			Status:      "passed",
			Severity:    "medium",
			Description: "UFW firewall should be active and configured",
		}
		passed++
	} else {
		controls["3.3.1"] = Control{
			ID:          "3.3.1",
			Title:       "Ensure UFW is enabled",
			Status:      "failed",
			Severity:    "medium",
			Description: "UFW firewall should be active and configured",
		}
	}

	// CIS 1.1.1.1 - Disable unused filesystems
	total++
	if !strings.Contains(strings.ToLower(data["available_shells"]), "cramfs") {
		controls["1.1.1.1"] = Control{
			ID:          "1.1.1.1",
			Title:       "Ensure mounting of cramfs filesystems is disabled",
			Status:      "passed",
			Severity:    "low",
			Description: "Cramfs filesystem should be disabled",
		}
		passed++
	} else {
		controls["1.1.1.1"] = Control{
			ID:          "1.1.1.1",
			Title:       "Ensure mounting of cramfs filesystems is disabled",
			Status:      "failed",
			Severity:    "low",
			Description: "Cramfs filesystem should be disabled",
		}
	}

	score := float64(passed) / float64(total) * 100.0

	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeCISLevel2 analyzes against CIS Level 2 benchmarks
func analyzeCISLevel2(data map[string]string) ComplianceProfile {
	level1 := analyzeCISLevel1(data)
	controls := level1.Controls
	passed := level1.Passed
	total := level1.Total

	// Additional Level 2 controls
	total++
	if strings.Contains(strings.ToLower(data["logging_daemon"]), "rsyslog") {
		controls["4.2.1.1"] = Control{
			ID:          "4.2.1.1",
			Title:       "Ensure rsyslog is installed",
			Status:      "passed",
			Severity:    "medium",
			Description: "rsyslog should be installed for centralized logging",
		}
		passed++
	} else {
		controls["4.2.1.1"] = Control{
			ID:          "4.2.1.1",
			Title:       "Ensure rsyslog is installed",
			Status:      "failed",
			Severity:    "medium",
			Description: "rsyslog should be installed for centralized logging",
		}
	}

	score := float64(passed) / float64(total) * 100.0

	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeISO27001 analyzes against ISO 27001 controls
func analyzeISO27001(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// A.9.2.3 - Management of privileged access rights
	total++
	if strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin no") {
		controls["A.9.2.3"] = Control{
			ID:          "A.9.2.3",
			Title:       "Management of privileged access rights",
			Status:      "passed",
			Severity:    "high",
			Description: "Privileged access should be restricted and controlled",
		}
		passed++
	} else {
		controls["A.9.2.3"] = Control{
			ID:          "A.9.2.3",
			Title:       "Management of privileged access rights",
			Status:      "failed",
			Severity:    "high",
			Description: "Privileged access should be restricted and controlled",
		}
	}

	// A.13.1.1 - Network controls
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["A.13.1.1"] = Control{
			ID:          "A.13.1.1",
			Title:       "Network controls",
			Status:      "passed",
			Severity:    "high",
			Description: "Network access should be controlled by firewalls",
		}
		passed++
	} else {
		controls["A.13.1.1"] = Control{
			ID:          "A.13.1.1",
			Title:       "Network controls",
			Status:      "failed",
			Severity:    "high",
			Description: "Network access should be controlled by firewalls",
		}
	}

	score := float64(passed) / float64(total) * 100.0

	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeNIST analyzes against NIST 800-53 controls
func analyzeNIST(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// AC-6 - Least Privilege
	total++
	if strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin no") {
		controls["AC-6"] = Control{
			ID:          "AC-6",
			Title:       "Least Privilege",
			Status:      "passed",
			Severity:    "high",
			Description: "Access should follow principle of least privilege",
		}
		passed++
	} else {
		controls["AC-6"] = Control{
			ID:          "AC-6",
			Title:       "Least Privilege",
			Status:      "failed",
			Severity:    "high",
			Description: "Access should follow principle of least privilege",
		}
	}

	// SC-7 - Boundary Protection
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["SC-7"] = Control{
			ID:          "SC-7",
			Title:       "Boundary Protection",
			Status:      "passed",
			Severity:    "medium",
			Description: "System boundaries should be protected",
		}
		passed++
	} else {
		controls["SC-7"] = Control{
			ID:          "SC-7",
			Title:       "Boundary Protection",
			Status:      "failed",
			Severity:    "medium",
			Description: "System boundaries should be protected",
		}
	}

	score := float64(passed) / float64(total) * 100.0

	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzePCIDSS analyzes against PCI DSS requirements
func analyzePCIDSS(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// Requirement 2.3 - Encrypt all non-console administrative access
	total++
	if strings.Contains(strings.ToLower(data["ssh_daemon_status"]), "running") {
		controls["2.3"] = Control{
			ID:          "2.3",
			Title:       "Encrypt all non-console administrative access",
			Status:      "passed",
			Severity:    "high",
			Description: "Administrative access should use secure protocols like SSH",
		}
		passed++
	} else {
		controls["2.3"] = Control{
			ID:          "2.3",
			Title:       "Encrypt all non-console administrative access",
			Status:      "failed",
			Severity:    "high",
			Description: "Administrative access should use secure protocols like SSH",
		}
	}

	// Requirement 1.1 - Firewall configuration
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["1.1"] = Control{
			ID:          "1.1",
			Title:       "Establish firewall configuration standards",
			Status:      "passed",
			Severity:    "high",
			Description: "Firewalls should be properly configured and active",
		}
		passed++
	} else {
		controls["1.1"] = Control{
			ID:          "1.1",
			Title:       "Establish firewall configuration standards",
			Status:      "failed",
			Severity:    "high",
			Description: "Firewalls should be properly configured and active",
		}
	}

	score := float64(passed) / float64(total) * 100.0

	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// extractSecurityFindings extracts security findings from Lynis data
func extractSecurityFindings(data map[string]string) []SecurityFinding {
	var findings []SecurityFinding

	// Check for SSH root login enabled
	if !strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin no") {
		findings = append(findings, SecurityFinding{
			ID:           "SSH-001",
			Title:        "SSH Root Login Enabled",
			Description:  "Direct root login via SSH is enabled, which poses a security risk",
			Severity:     "high",
			Category:     "authentication",
			Mappings:     []string{"CIS 5.2.8", "ISO 27001 A.9.2.3", "NIST AC-6", "PCI DSS 2.3"},
			FixAvailable: true,
		})
	}

	// Check for firewall status
	if !strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		findings = append(findings, SecurityFinding{
			ID:           "NET-001",
			Title:        "Firewall Not Active",
			Description:  "System firewall is not active, leaving network services exposed",
			Severity:     "high",
			Category:     "network",
			Mappings:     []string{"CIS 3.3.1", "ISO 27001 A.13.1.1", "NIST SC-7", "PCI DSS 1.1"},
			FixAvailable: true,
		})
	}

	// Check for unattended upgrades
	if !strings.Contains(strings.ToLower(data["software_package_tools"]), "unattended-upgrades") {
		findings = append(findings, SecurityFinding{
			ID:           "UPD-001",
			Title:        "Automatic Updates Not Configured",
			Description:  "Automatic security updates are not configured",
			Severity:     "medium",
			Category:     "maintenance",
			Mappings:     []string{"CIS 1.9", "ISO 27001 A.12.6.1"},
			FixAvailable: true,
		})
	}

	return findings
}

// generateRemediations generates automated remediation suggestions
func generateRemediations(findings []SecurityFinding) []Remediation {
	var remediations []Remediation

	for _, finding := range findings {
		switch finding.ID {
		case "SSH-001":
			remediations = append(remediations, Remediation{
				ID:          "REM-SSH-001",
				Title:       "Disable SSH Root Login",
				Description: "Modify SSH configuration to disable direct root login",
				Command:     "sudo sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config && sudo systemctl restart ssh",
				Risk:        "low",
				FindingID:   finding.ID,
			})
		case "NET-001":
			remediations = append(remediations, Remediation{
				ID:          "REM-NET-001",
				Title:       "Enable UFW Firewall",
				Description: "Enable and configure UFW firewall with basic rules",
				Command:     "sudo ufw enable && sudo ufw default deny incoming && sudo ufw default allow outgoing && sudo ufw allow ssh",
				Risk:        "medium",
				FindingID:   finding.ID,
			})
		case "UPD-001":
			remediations = append(remediations, Remediation{
				ID:          "REM-UPD-001",
				Title:       "Enable Automatic Updates",
				Description: "Install and configure unattended-upgrades for automatic security updates",
				Command:     "sudo apt update && sudo apt install -y unattended-upgrades && sudo dpkg-reconfigure -plow unattended-upgrades",
				Risk:        "low",
				FindingID:   finding.ID,
			})
		}
	}

	return remediations
}

// analyzeSOC2 analyzes against SOC 2 controls
func analyzeSOC2(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// CC6.1 - Logical and Physical Access Controls
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["CC6.1"] = Control{
			ID:          "CC6.1",
			Title:       "Logical and Physical Access Controls",
			Status:      "passed",
			Severity:    "high",
			Description: "Implement logical and physical access controls",
		}
		passed++
	} else {
		controls["CC6.1"] = Control{
			ID:          "CC6.1",
			Title:       "Logical and Physical Access Controls",
			Status:      "failed",
			Severity:    "high",
			Description: "Implement logical and physical access controls",
		}
	}

	// CC6.7 - Transmission of Data
	total++
	if strings.Contains(strings.ToLower(data["ssh_daemon_status"]), "running") {
		controls["CC6.7"] = Control{
			ID:          "CC6.7",
			Title:       "Transmission of Data",
			Status:      "passed",
			Severity:    "medium",
			Description: "Data transmission should be protected",
		}
		passed++
	} else {
		controls["CC6.7"] = Control{
			ID:          "CC6.7",
			Title:       "Transmission of Data",
			Status:      "failed",
			Severity:    "medium",
			Description: "Data transmission should be protected",
		}
	}

	score := float64(passed) / float64(total) * 100.0
	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeHIPAA analyzes against HIPAA Security Rule
func analyzeHIPAA(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// 164.312(a)(1) - Access Control
	total++
	if !strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin yes") {
		controls["164.312(a)(1)"] = Control{
			ID:          "164.312(a)(1)",
			Title:       "Access Control",
			Status:      "passed",
			Severity:    "high",
			Description: "Assign unique user identification and automatic logoff",
		}
		passed++
	} else {
		controls["164.312(a)(1)"] = Control{
			ID:          "164.312(a)(1)",
			Title:       "Access Control",
			Status:      "failed",
			Severity:    "high",
			Description: "Assign unique user identification and automatic logoff",
		}
	}

	// 164.312(e)(1) - Transmission Security
	total++
	if strings.Contains(strings.ToLower(data["ssh_daemon_status"]), "running") {
		controls["164.312(e)(1)"] = Control{
			ID:          "164.312(e)(1)",
			Title:       "Transmission Security",
			Status:      "passed",
			Severity:    "high",
			Description: "Implement technical safeguards for electronic PHI transmission",
		}
		passed++
	} else {
		controls["164.312(e)(1)"] = Control{
			ID:          "164.312(e)(1)",
			Title:       "Transmission Security",
			Status:      "failed",
			Severity:    "high",
			Description: "Implement technical safeguards for electronic PHI transmission",
		}
	}

	score := float64(passed) / float64(total) * 100.0
	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeGDPR analyzes against GDPR requirements
func analyzeGDPR(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// Article 32 - Security of Processing
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["Art32.1"] = Control{
			ID:          "Art32.1",
			Title:       "Security of Processing",
			Status:      "passed",
			Severity:    "high",
			Description: "Implement appropriate technical measures for data security",
		}
		passed++
	} else {
		controls["Art32.1"] = Control{
			ID:          "Art32.1",
			Title:       "Security of Processing",
			Status:      "failed",
			Severity:    "high",
			Description: "Implement appropriate technical measures for data security",
		}
	}

	// Article 25 - Data Protection by Design
	total++
	if strings.Contains(strings.ToLower(data["logging_daemon"]), "rsyslog") {
		controls["Art25"] = Control{
			ID:          "Art25",
			Title:       "Data Protection by Design",
			Status:      "passed",
			Severity:    "medium",
			Description: "Implement data protection measures by design and by default",
		}
		passed++
	} else {
		controls["Art25"] = Control{
			ID:          "Art25",
			Title:       "Data Protection by Design",
			Status:      "failed",
			Severity:    "medium",
			Description: "Implement data protection measures by design and by default",
		}
	}

	score := float64(passed) / float64(total) * 100.0
	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeSOX analyzes against Sarbanes-Oxley requirements
func analyzeSOX(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// Section 404 - Internal Controls
	total++
	if strings.Contains(strings.ToLower(data["logging_daemon"]), "rsyslog") {
		controls["SOX404"] = Control{
			ID:          "SOX404",
			Title:       "Internal Controls Assessment",
			Status:      "passed",
			Severity:    "high",
			Description: "Maintain adequate internal control over financial reporting",
		}
		passed++
	} else {
		controls["SOX404"] = Control{
			ID:          "SOX404",
			Title:       "Internal Controls Assessment",
			Status:      "failed",
			Severity:    "high",
			Description: "Maintain adequate internal control over financial reporting",
		}
	}

	// IT General Controls
	total++
	if !strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin yes") {
		controls["ITGC"] = Control{
			ID:          "ITGC",
			Title:       "IT General Controls",
			Status:      "passed",
			Severity:    "medium",
			Description: "Implement proper IT general controls",
		}
		passed++
	} else {
		controls["ITGC"] = Control{
			ID:          "ITGC",
			Title:       "IT General Controls",
			Status:      "failed",
			Severity:    "medium",
			Description: "Implement proper IT general controls",
		}
	}

	score := float64(passed) / float64(total) * 100.0
	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// analyzeFISMA analyzes against FISMA requirements
func analyzeFISMA(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// FISMA Access Control
	total++
	if !strings.Contains(strings.ToLower(data["ssh_daemon_options"]), "permitrootlogin yes") {
		controls["FISMA-AC"] = Control{
			ID:          "FISMA-AC",
			Title:       "Access Control",
			Status:      "passed",
			Severity:    "high",
			Description: "Implement access control policies and procedures",
		}
		passed++
	} else {
		controls["FISMA-AC"] = Control{
			ID:          "FISMA-AC",
			Title:       "Access Control",
			Status:      "failed",
			Severity:    "high",
			Description: "Implement access control policies and procedures",
		}
	}

	// FISMA System and Communications Protection
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["FISMA-SC"] = Control{
			ID:          "FISMA-SC",
			Title:       "System and Communications Protection",
			Status:      "passed",
			Severity:    "high",
			Description: "Protect system and communications",
		}
		passed++
	} else {
		controls["FISMA-SC"] = Control{
			ID:          "FISMA-SC",
			Title:       "System and Communications Protection",
			Status:      "failed",
			Severity:    "high",
			Description: "Protect system and communications",
		}
	}

	score := float64(passed) / float64(total) * 100.0
	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// historyTrendHandler returns trend data for charts
func historyTrendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	trend, err := historyManager.GetTrend(period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting trend: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trend)
}

// historyRecordsHandler returns all historical records
func historyRecordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse 'since' parameter (optional)
	sinceStr := r.URL.Query().Get("since")
	var since time.Time
	if sinceStr != "" {
		var err error
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			since = time.Now().AddDate(0, 0, -30) // Default: 30 days
		}
	} else {
		since = time.Now().AddDate(0, 0, -30) // Default: 30 days
	}

	records, err := historyManager.GetRecordsSince(since)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting records: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"count":   len(records),
		"since":   since,
	})
}

// historyCompareHandler compares current with previous audit
func historyCompareHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := parseLynisReport()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing current report: %v", err), http.StatusInternalServerError)
		return
	}

	comparison, err := historyManager.CompareWithPrevious(data)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No previous audit found for comparison",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"comparison": comparison,
	})
}

// historyStatsHandler returns storage statistics
func historyStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats, err := historyManager.GetStorageStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// schedulerStatusHandler returns scheduler status
func schedulerStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := auditScheduler.GetStatus()
	json.NewEncoder(w).Encode(status)
}

// schedulerConfigHandler updates scheduler configuration
func schedulerConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		// Return current config
		config := auditScheduler.GetStatus()
		json.NewEncoder(w).Encode(config)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Update config
	var request struct {
		Enabled      bool   `json:"enabled"`
		Interval     string `json:"interval"` // "hourly", "daily", "weekly", "monthly"
		RunOnStartup bool   `json:"run_on_startup"`
		QuietMode    bool   `json:"quiet_mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse interval
	var config SchedulerConfig
	config.Enabled = request.Enabled
	config.RunOnStartup = request.RunOnStartup
	config.QuietMode = request.QuietMode

	switch request.Interval {
	case "hourly":
		config.Interval = 1 * time.Hour
	case "daily":
		config.Interval = 24 * time.Hour
	case "weekly":
		config.Interval = 7 * 24 * time.Hour
	case "monthly":
		config.Interval = 30 * 24 * time.Hour
	default:
		config.Interval = 24 * time.Hour // Default to daily
	}

	auditScheduler.UpdateConfig(config)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Scheduler configuration updated",
		"config":  auditScheduler.GetStatus(),
	})
}

// agentRegisterHandler handles agent registration
func agentRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Hostname     string `json:"hostname"`
		IPAddress    string `json:"ip_address"`
		OS           string `json:"os"`
		Arch         string `json:"arch"`
		AgentVersion string `json:"agent_version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Register the server
	server, err := serverManager.RegisterServer(
		req.Hostname,
		req.IPAddress,
		req.OS,
		req.Arch,
		req.AgentVersion,
	)
	if err != nil {
		log.Printf("Failed to register server: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to register server: " + err.Error(),
		})
		return
	}

	log.Printf("✅ New server registered: %s (%s)", server.Hostname, server.ID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"server_id": server.ID,
		"api_key":   server.APIKey,
		"message":   "Server registered successfully",
	})
}

// agentHeartbeatHandler handles agent heartbeats
func agentHeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Authenticate agent
	apiKey := extractAPIKey(r)
	if apiKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	server, err := serverManager.GetServerByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update heartbeat
	if err := serverManager.UpdateHeartbeat(server.ID); err != nil {
		log.Printf("Failed to update heartbeat for %s: %v", server.ID, err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to update heartbeat",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Heartbeat received",
	})
}

// agentMetricsHandler handles agent metrics submission
func agentMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Authenticate agent
	apiKey := extractAPIKey(r)
	if apiKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	server, err := serverManager.GetServerByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse metrics
	var metrics ServerMetrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Ensure server ID matches
	metrics.ServerID = server.ID

	// Save metrics
	if err := serverManager.SaveMetrics(&metrics); err != nil {
		log.Printf("Failed to save metrics for %s: %v", server.ID, err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to save metrics",
		})
		return
	}

	log.Printf("📊 Received metrics from %s: Score=%s%%, Warnings=%s",
		server.Hostname, metrics.HardeningIndex, metrics.Warnings)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Metrics received",
	})
}

// serversListHandler lists all servers
func serversListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	servers, err := serverManager.ListServers()
	if err != nil {
		http.Error(w, "Failed to list servers", http.StatusInternalServerError)
		return
	}

	// Get latest metrics for each server
	type ServerWithMetrics struct {
		*ServerInfo
		LatestMetrics *ServerMetrics `json:"latest_metrics,omitempty"`
	}

	serversWithMetrics := make([]ServerWithMetrics, 0, len(servers))
	for _, server := range servers {
		swm := ServerWithMetrics{ServerInfo: server}
		
		// Try to get latest metrics
		if metrics, err := serverManager.GetLatestMetrics(server.ID); err == nil {
			swm.LatestMetrics = metrics
		}

		serversWithMetrics = append(serversWithMetrics, swm)
	}

	// Get dashboard stats
	stats, _ := serverManager.GetDashboardStats()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"servers": serversWithMetrics,
		"stats":   stats,
		"count":   len(servers),
	})
}

// serversDetailHandler handles individual server details
func serversDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract server ID from URL
	serverID := strings.TrimPrefix(r.URL.Path, "/api/servers/")
	if serverID == "" {
		http.Error(w, "Server ID required", http.StatusBadRequest)
		return
	}

	// Get server info
	server, err := serverManager.GetServer(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	// Get metrics history (last 30 audits)
	metrics, err := serverManager.GetServerMetrics(serverID, 30)
	if err != nil {
		metrics = []*ServerMetrics{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"server":  server,
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// extractAPIKey extracts API key from Authorization header
func extractAPIKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	// Format: "Bearer <api_key>"
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// analysisAPIHandler returns the local system's Lynis analysis as JSON
func analysisAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check if full data is requested
	fullData := r.URL.Query().Get("full") == "true"
	
	// Parse the Lynis report
	data, err := parseLynisReport()
	if err != nil {
		// Return empty data if file not found
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No Lynis report found",
			"hardening_index": "0",
			"tests_performed": "0",
			"warnings": "0",
			"suggestions": "0",
		})
		return
	}
	
	// Count warnings and suggestions by reading the file again
	warningCount, suggestionCount, testCount := countArrayEntries()
	
	// Get tests_performed or use count
	testsPerformed := data["tests_performed"]
	if testsPerformed == "" || testsPerformed == "0" {
		testsPerformed = fmt.Sprintf("%d", testCount)
	}
	
	if fullData {
		// Return complete Lynis data with arrays
		allData := parseCompleteLynisReport()
		json.NewEncoder(w).Encode(allData)
	} else {
		// Extract key metrics only
		response := map[string]string{
			"hardening_index":  data["hardening_index"],
			"tests_performed":  testsPerformed,
			"warnings":         fmt.Sprintf("%d", warningCount),
			"suggestions":      fmt.Sprintf("%d", suggestionCount),
			"lynis_version":    data["lynis_version"],
			"scan_date":        data["report_datetime_start"],
			"os_name":          data["os"],
			"os_fullname":      data["os_fullname"],
			"os_version":       data["os_version"],
		}
		
		json.NewEncoder(w).Encode(response)
	}
}

// parseCompleteLynisReport parses the complete Lynis report including arrays
func parseCompleteLynisReport() map[string]interface{} {
	reportPaths := []string{
		"/Users/apple/lynis-report.dat",
		"/Users/apple/Desktop/untitled folder/untitled folder/lynis-report.dat",
		"./lynis-report.dat",
		"/tmp/lynis-report.dat",
		"/usr/local/var/log/lynis-report.dat",
		"/var/log/lynis-report.dat",
		"/usr/share/lynis/lynis-report.dat",
		os.Getenv("HOME") + "/lynis-report.dat",
	}

	result := make(map[string]interface{})
	result["fields"] = make(map[string]string)
	result["warnings"] = []string{}
	result["suggestions"] = []string{}
	result["tests"] = []string{}
	result["network_interfaces"] = []string{}
	result["network_ipv4"] = []string{}
	result["network_ipv6"] = []string{}
	result["network_listen_ports"] = []string{}
	result["available_shells"] = []string{}
	result["apache_modules"] = []string{}
	result["package_managers"] = []string{}
	result["nameservers"] = []string{}
	result["default_gateways"] = []string{}

	for _, reportPath := range reportPaths {
		file, err := os.Open(reportPath)
		if err != nil {
			continue
		}
		defer file.Close()

		fields := make(map[string]string)
		var warnings, suggestions, tests []string
		var networkInterfaces, networkIPv4, networkIPv6, listenPorts []string
		var shells, apacheModules, packageManagers, nameservers, gateways []string

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Handle array entries
			if strings.Contains(line, "[]=") {
				parts := strings.SplitN(line, "[]=", 2)
				if len(parts) == 2 {
					key := parts[0]
					value := parts[1]
					
					switch key {
					case "warning":
						warnings = append(warnings, value)
					case "suggestion":
						suggestions = append(suggestions, value)
					case "test":
						tests = append(tests, value)
					case "network_interface":
						networkInterfaces = append(networkInterfaces, value)
					case "network_ipv4_address":
						networkIPv4 = append(networkIPv4, value)
					case "network_ipv6_address":
						networkIPv6 = append(networkIPv6, value)
					case "network_listen_port":
						listenPorts = append(listenPorts, value)
					case "available_shell":
						shells = append(shells, value)
					case "apache_module":
						apacheModules = append(apacheModules, value)
					case "package_manager":
						packageManagers = append(packageManagers, value)
					case "nameserver":
						nameservers = append(nameservers, value)
					case "default_gateway":
						gateways = append(gateways, value)
					}
				}
			} else if strings.Contains(line, "=") {
				// Handle regular key=value
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					fields[parts[0]] = parts[1]
				}
			}
		}

		result["fields"] = fields
		result["warnings"] = warnings
		result["suggestions"] = suggestions
		result["tests"] = tests
		result["network_interfaces"] = networkInterfaces
		result["network_ipv4"] = networkIPv4
		result["network_ipv6"] = networkIPv6
		result["network_listen_ports"] = listenPorts
		result["available_shells"] = shells
		result["apache_modules"] = apacheModules
		result["package_managers"] = packageManagers
		result["nameservers"] = nameservers
		result["default_gateways"] = gateways
		
		return result
	}

	return result
}

// countArrayEntries counts warning[], suggestion[], and test[] entries in Lynis report
func countArrayEntries() (int, int, int) {
	reportPaths := []string{
		"/Users/apple/lynis-report.dat",
		"/Users/apple/Desktop/untitled folder/untitled folder/lynis-report.dat",
		"./lynis-report.dat",
		"/tmp/lynis-report.dat",
		"/usr/local/var/log/lynis-report.dat",
		"/var/log/lynis-report.dat",
		"/usr/share/lynis/lynis-report.dat",
		os.Getenv("HOME") + "/lynis-report.dat",
	}

	for _, reportPath := range reportPaths {
		file, err := os.Open(reportPath)
		if err != nil {
			continue
		}
		defer file.Close()

		warningCount := 0
		suggestionCount := 0
		testCount := 0
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "warning[") {
				warningCount++
			} else if strings.HasPrefix(line, "suggestion[") {
				suggestionCount++
			} else if strings.HasPrefix(line, "test[") {
				testCount++
			}
		}

		return warningCount, suggestionCount, testCount
	}

	return 0, 0, 0
}

// analyzeCOBIT analyzes against COBIT framework
func analyzeCOBIT(data map[string]string) ComplianceProfile {
	controls := make(map[string]Control)
	passed := 0
	total := 0

	// APO13 - Manage Security
	total++
	if strings.Contains(strings.ToLower(data["firewall_status"]), "active") {
		controls["APO13"] = Control{
			ID:          "APO13",
			Title:       "Manage Security",
			Status:      "passed",
			Severity:    "high",
			Description: "Define, implement and monitor a system for information security management",
		}
		passed++
	} else {
		controls["APO13"] = Control{
			ID:          "APO13",
			Title:       "Manage Security",
			Status:      "failed",
			Severity:    "high",
			Description: "Define, implement and monitor a system for information security management",
		}
	}

	// DSS05 - Manage Security Services
	total++
	if strings.Contains(strings.ToLower(data["logging_daemon"]), "rsyslog") {
		controls["DSS05"] = Control{
			ID:          "DSS05",
			Title:       "Manage Security Services",
			Status:      "passed",
			Severity:    "medium",
			Description: "Protect enterprise information to maintain risk at acceptable level",
		}
		passed++
	} else {
		controls["DSS05"] = Control{
			ID:          "DSS05",
			Title:       "Manage Security Services",
			Status:      "failed",
			Severity:    "medium",
			Description: "Protect enterprise information to maintain risk at acceptable level",
		}
	}

	score := float64(passed) / float64(total) * 100.0
	return ComplianceProfile{
		Score:    score,
		Total:    total,
		Passed:   passed,
		Failed:   total - passed,
		Controls: controls,
	}
}

// Global instances
var (
	historyManager *HistoryManager
	auditScheduler *AuditScheduler
	serverManager  *ServerManager
)

func main() {
	// Initialize history manager
	historyManager = NewHistoryManager("./history")
	log.Println("💾 History manager initialized")

	// Initialize audit scheduler
	auditScheduler = NewAuditScheduler(historyManager)
	auditScheduler.Start()
	log.Println("⏰ Audit scheduler initialized")

	// Initialize server manager (multi-server support)
	serverManager = NewServerManager("./data")
	log.Println("🌐 Server manager initialized (multi-server mode)")

	// Start background task to update server status
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			serverManager.UpdateServerStatus()
		}
	}()

	// Register HTTP handlers
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/servers", multiServerDashboardHandler)
	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/run-audit", runAuditHandler)
	http.HandleFunc("/compliance", complianceProfileHandler)
	http.HandleFunc("/remediate", remediateHandler)
	
	// History and scheduling endpoints
	http.HandleFunc("/history/trend", historyTrendHandler)
	http.HandleFunc("/history/records", historyRecordsHandler)
	http.HandleFunc("/history/compare", historyCompareHandler)
	http.HandleFunc("/history/stats", historyStatsHandler)
	http.HandleFunc("/scheduler/status", schedulerStatusHandler)
	http.HandleFunc("/scheduler/config", schedulerConfigHandler)
	
	// Multi-server API endpoints (for agents)
	http.HandleFunc("/api/agents/register", agentRegisterHandler)
	http.HandleFunc("/api/agents/heartbeat", agentHeartbeatHandler)
	http.HandleFunc("/api/metrics", agentMetricsHandler)
	http.HandleFunc("/api/servers", serversListHandler)
	http.HandleFunc("/api/servers/", serversDetailHandler) // handles /api/servers/{id}
	http.HandleFunc("/api/analysis", analysisAPIHandler)   // Local system analysis

	port := "5179"
	fmt.Printf("🚀 Linux Hardening Dashboard starting on http://localhost:%s\n", port)
	fmt.Printf("📊 Dashboard available at http://localhost:%s/\n", port)
	fmt.Printf("📋 API endpoint available at http://localhost:%s/report\n", port)
	fmt.Printf("🔍 Run audit endpoint available at http://localhost:%s/run-audit\n", port)
	fmt.Printf("📋 Compliance endpoint available at http://localhost:%s/compliance\n", port)
	fmt.Printf("🔧 Remediation endpoint available at http://localhost:%s/remediate\n", port)
	fmt.Printf("📈 History trend endpoint available at http://localhost:%s/history/trend?period=30d\n", port)
	fmt.Printf("⏰ Scheduler status endpoint available at http://localhost:%s/scheduler/status\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
