package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ServerInfo represents a registered server
type ServerInfo struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	OS           string    `json:"os"`
	Arch         string    `json:"arch"`
	AgentVersion string    `json:"agent_version"`
	APIKey       string    `json:"api_key"`
	Status       string    `json:"status"` // active, warning, offline
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ServerMetrics represents metrics from a server
type ServerMetrics struct {
	ServerID        string                 `json:"server_id"`
	Timestamp       time.Time              `json:"timestamp"`
	HardeningIndex  string                 `json:"hardening_index"`
	Warnings        string                 `json:"warnings"`
	TestsPerformed  string                 `json:"tests_performed"`
	ComplianceScore map[string]interface{} `json:"compliance_score"`
	RawData         map[string]string      `json:"raw_data"`
}

// ServerManager manages multiple servers
type ServerManager struct {
	serversDir string
	mu         sync.RWMutex
}

// NewServerManager creates a new server manager
func NewServerManager(dataDir string) *ServerManager {
	serversDir := filepath.Join(dataDir, "servers")
	os.MkdirAll(serversDir, 0755)

	return &ServerManager{
		serversDir: serversDir,
	}
}

// RegisterServer registers a new server
func (sm *ServerManager) RegisterServer(hostname, ipAddress, osName, arch, agentVersion string) (*ServerInfo, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Generate unique ID and API key
	id := generateID()
	apiKey := generateAPIKey()

	server := &ServerInfo{
		ID:           id,
		Hostname:     hostname,
		IPAddress:    ipAddress,
		OS:           osName,
		Arch:         arch,
		AgentVersion: agentVersion,
		APIKey:       apiKey,
		Status:       "active",
		LastHeartbeat: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create server directory
	serverDir := filepath.Join(sm.serversDir, id)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create server directory: %w", err)
	}

	// Create audits subdirectory
	auditsDir := filepath.Join(serverDir, "audits")
	if err := os.MkdirAll(auditsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audits directory: %w", err)
	}

	// Save server info
	if err := sm.saveServerInfo(server); err != nil {
		return nil, err
	}

	return server, nil
}

// UpdateHeartbeat updates server's last heartbeat
func (sm *ServerManager) UpdateHeartbeat(serverID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	server, err := sm.loadServerInfo(serverID)
	if err != nil {
		return err
	}

	server.LastHeartbeat = time.Now()
	server.UpdatedAt = time.Now()
	server.Status = "active"

	return sm.saveServerInfo(server)
}

// SaveMetrics saves audit metrics for a server
func (sm *ServerManager) SaveMetrics(metrics *ServerMetrics) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Verify server exists
	if _, err := sm.loadServerInfo(metrics.ServerID); err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Save metrics to file
	auditsDir := filepath.Join(sm.serversDir, metrics.ServerID, "audits")
	filename := fmt.Sprintf("audit_%s.json", metrics.Timestamp.Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(auditsDir, filename)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// GetServer returns server info by ID
func (sm *ServerManager) GetServer(serverID string) (*ServerInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.loadServerInfo(serverID)
}

// GetServerByAPIKey returns server info by API key
func (sm *ServerManager) GetServerByAPIKey(apiKey string) (*ServerInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	servers, err := sm.listServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.APIKey == apiKey {
			return server, nil
		}
	}

	return nil, fmt.Errorf("server not found with API key")
}

// ListServers returns all registered servers
func (sm *ServerManager) ListServers() ([]*ServerInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.listServers()
}

// GetServerMetrics returns recent metrics for a server
func (sm *ServerManager) GetServerMetrics(serverID string, limit int) ([]*ServerMetrics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	auditsDir := filepath.Join(sm.serversDir, serverID, "audits")
	
	files, err := os.ReadDir(auditsDir)
	if err != nil {
		return nil, err
	}

	var metrics []*ServerMetrics
	count := 0

	// Read files in reverse order (newest first)
	for i := len(files) - 1; i >= 0 && count < limit; i-- {
		file := files[i]
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(auditsDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var m ServerMetrics
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}

		metrics = append(metrics, &m)
		count++
	}

	return metrics, nil
}

// GetLatestMetrics returns the most recent metrics for a server
func (sm *ServerManager) GetLatestMetrics(serverID string) (*ServerMetrics, error) {
	metrics, err := sm.GetServerMetrics(serverID, 1)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no metrics found")
	}

	return metrics[0], nil
}

// UpdateServerStatus updates server status based on heartbeat
func (sm *ServerManager) UpdateServerStatus() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	servers, err := sm.listServers()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, server := range servers {
		timeSinceHeartbeat := now.Sub(server.LastHeartbeat)

		if timeSinceHeartbeat > 10*time.Minute {
			server.Status = "offline"
		} else if timeSinceHeartbeat > 7*time.Minute {
			server.Status = "warning"
		} else {
			server.Status = "active"
		}

		sm.saveServerInfo(server)
	}

	return nil
}

// GetDashboardStats returns overall statistics
func (sm *ServerManager) GetDashboardStats() (map[string]interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	servers, err := sm.listServers()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_servers": len(servers),
		"active":        0,
		"warning":       0,
		"offline":       0,
		"avg_score":     0.0,
		"total_warnings": 0,
	}

	totalScore := 0.0
	serverCount := 0

	for _, server := range servers {
		switch server.Status {
		case "active":
			stats["active"] = stats["active"].(int) + 1
		case "warning":
			stats["warning"] = stats["warning"].(int) + 1
		case "offline":
			stats["offline"] = stats["offline"].(int) + 1
		}

		// Get latest metrics for score
		metrics, err := sm.GetLatestMetrics(server.ID)
		if err == nil && metrics.HardeningIndex != "" {
			var score float64
			fmt.Sscanf(metrics.HardeningIndex, "%f", &score)
			totalScore += score
			serverCount++

			var warnings int
			fmt.Sscanf(metrics.Warnings, "%d", &warnings)
			stats["total_warnings"] = stats["total_warnings"].(int) + warnings
		}
	}

	if serverCount > 0 {
		stats["avg_score"] = totalScore / float64(serverCount)
	}

	return stats, nil
}

// DeleteServer removes a server and all its data
func (sm *ServerManager) DeleteServer(serverID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	serverDir := filepath.Join(sm.serversDir, serverID)
	return os.RemoveAll(serverDir)
}

// Helper functions

func (sm *ServerManager) saveServerInfo(server *ServerInfo) error {
	serverDir := filepath.Join(sm.serversDir, server.ID)
	infoPath := filepath.Join(serverDir, "info.json")

	data, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(infoPath, data, 0644)
}

func (sm *ServerManager) loadServerInfo(serverID string) (*ServerInfo, error) {
	infoPath := filepath.Join(sm.serversDir, serverID, "info.json")

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	var server ServerInfo
	if err := json.Unmarshal(data, &server); err != nil {
		return nil, err
	}

	return &server, nil
}

func (sm *ServerManager) listServers() ([]*ServerInfo, error) {
	entries, err := os.ReadDir(sm.serversDir)
	if err != nil {
		return nil, err
	}

	var servers []*ServerInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		server, err := sm.loadServerInfo(entry.Name())
		if err != nil {
			continue
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

