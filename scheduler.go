package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// SchedulerConfig holds scheduling configuration
type SchedulerConfig struct {
	Enabled       bool
	Interval      time.Duration
	RunOnStartup  bool
	QuietMode     bool
	LastRunTime   time.Time
	NextRunTime   time.Time
}

// AuditScheduler manages scheduled Lynis audits
type AuditScheduler struct {
	config         SchedulerConfig
	historyManager *HistoryManager
	stopChan       chan bool
	running        bool
}

// NewAuditScheduler creates a new scheduler
func NewAuditScheduler(historyManager *HistoryManager) *AuditScheduler {
	return &AuditScheduler{
		config: SchedulerConfig{
			Enabled:      false, // Disabled by default, user can enable via settings
			Interval:     24 * time.Hour, // Default: daily
			RunOnStartup: false,
			QuietMode:    true,
		},
		historyManager: historyManager,
		stopChan:       make(chan bool),
		running:        false,
	}
}

// Start begins the scheduler
func (s *AuditScheduler) Start() error {
	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	if !s.config.Enabled {
		log.Println("⏰ Scheduler is disabled. Enable it in settings to run automated audits.")
		return nil
	}

	s.running = true
	log.Printf("⏰ Audit scheduler started - will run every %s\n", s.config.Interval)

	// Run on startup if configured
	if s.config.RunOnStartup {
		log.Println("🚀 Running audit on startup...")
		go s.runAudit()
	}

	// Start the scheduling loop
	go s.scheduleLoop()

	return nil
}

// Stop stops the scheduler
func (s *AuditScheduler) Stop() {
	if !s.running {
		return
	}

	log.Println("⏰ Stopping audit scheduler...")
	s.stopChan <- true
	s.running = false
}

// UpdateConfig updates scheduler configuration
func (s *AuditScheduler) UpdateConfig(config SchedulerConfig) {
	wasRunning := s.running
	
	if wasRunning {
		s.Stop()
	}

	s.config = config

	if wasRunning && config.Enabled {
		s.Start()
	}
}

// GetStatus returns current scheduler status
func (s *AuditScheduler) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":       s.config.Enabled,
		"running":       s.running,
		"interval":      s.config.Interval.String(),
		"last_run":      s.config.LastRunTime,
		"next_run":      s.config.NextRunTime,
		"run_on_startup": s.config.RunOnStartup,
		"quiet_mode":    s.config.QuietMode,
	}
}

// scheduleLoop is the main scheduling loop
func (s *AuditScheduler) scheduleLoop() {
	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	// Calculate next run time
	s.config.NextRunTime = time.Now().Add(s.config.Interval)
	log.Printf("📅 Next scheduled audit: %s\n", s.config.NextRunTime.Format("2006-01-02 15:04:05"))

	for {
		select {
		case <-ticker.C:
			log.Println("⏰ Scheduled audit triggered")
			s.runAudit()
			s.config.NextRunTime = time.Now().Add(s.config.Interval)

		case <-s.stopChan:
			log.Println("⏰ Scheduler stopped")
			return
		}
	}
}

// runAudit executes a Lynis audit
func (s *AuditScheduler) runAudit() {
	s.config.LastRunTime = time.Now()
	log.Printf("🔍 Starting scheduled Lynis audit at %s\n", s.config.LastRunTime.Format("2006-01-02 15:04:05"))

	// Check if Lynis is available
	lynisPath, err := exec.LookPath("lynis")
	if err != nil {
		log.Printf("❌ Lynis not found: %v\n", err)
		return
	}

	log.Printf("✅ Found Lynis at: %s\n", lynisPath)

	// Prepare audit command
	var cmd *exec.Cmd
	if s.config.QuietMode {
		cmd = exec.Command("sudo", "lynis", "audit", "system", "--quick", "--quiet")
	} else {
		cmd = exec.Command("sudo", "lynis", "audit", "system", "--quick")
	}

	// Run the audit
	log.Println("🚀 Executing Lynis audit...")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Printf("⚠️ Lynis audit completed with warnings: %v\n", err)
		if !s.config.QuietMode {
			log.Printf("Output: %s\n", string(output))
		}
	} else {
		log.Println("✅ Lynis audit completed successfully")
	}

	// Save to history
	log.Println("💾 Saving audit results to history...")
	data, err := parseLynisReport()
	if err != nil {
		log.Printf("❌ Failed to parse Lynis report: %v\n", err)
		return
	}

	// Analyze compliance
	compliance := analyzeCompliance(data)

	// Save to history
	if err := s.historyManager.SaveAudit(data, compliance); err != nil {
		log.Printf("❌ Failed to save to history: %v\n", err)
		return
	}

	log.Println("✅ Audit results saved to history")

	// Log summary
	if hardening, exists := data["hardening_index"]; exists {
		log.Printf("📊 Security Score: %s%%\n", hardening)
	}
	if warnings, exists := data["warnings"]; exists {
		log.Printf("⚠️ Warnings: %s\n", warnings)
	}
}

// RunManualAudit runs an audit manually (called from API)
func (s *AuditScheduler) RunManualAudit() error {
	if !s.running {
		// If scheduler not running, run audit directly
		go s.runAudit()
		return nil
	}

	// Trigger immediate audit
	go s.runAudit()
	return nil
}

// Preset schedule configurations
var (
	// HourlySchedule runs audit every hour
	HourlySchedule = SchedulerConfig{
		Enabled:      true,
		Interval:     1 * time.Hour,
		RunOnStartup: false,
		QuietMode:    true,
	}

	// DailySchedule runs audit once per day
	DailySchedule = SchedulerConfig{
		Enabled:      true,
		Interval:     24 * time.Hour,
		RunOnStartup: false,
		QuietMode:    true,
	}

	// WeeklySchedule runs audit once per week
	WeeklySchedule = SchedulerConfig{
		Enabled:      true,
		Interval:     7 * 24 * time.Hour,
		RunOnStartup: false,
		QuietMode:    true,
	}

	// MonthlySchedule runs audit once per month
	MonthlySchedule = SchedulerConfig{
		Enabled:      true,
		Interval:     30 * 24 * time.Hour,
		RunOnStartup: false,
		QuietMode:    true,
	}
)

