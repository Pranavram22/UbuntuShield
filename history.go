package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// HistoryConfig manages historical audit data
type HistoryConfig struct {
	StoragePath       string
	MaxFullRecords    int           // Keep last N full records
	RetentionDays     int           // Delete records older than N days
	CompressionAge    time.Duration // Compress records older than this
	EnableCompression bool
}

// AuditRecord represents a single audit snapshot
type AuditRecord struct {
	Timestamp        time.Time              `json:"timestamp"`
	HardeningIndex   string                 `json:"hardening_index"`
	Warnings         string                 `json:"warnings"`
	TestsPerformed   string                 `json:"tests_performed"`
	Suggestions      int                    `json:"suggestions"`
	ComplianceScores map[string]float64     `json:"compliance_scores"`
	KeyMetrics       map[string]string      `json:"key_metrics"` // Store only important fields
	FullDataHash     string                 `json:"full_data_hash"`
	Compressed       bool                   `json:"compressed"`
}

// TrendData represents trend analysis over time
type TrendData struct {
	SecurityScoreTrend []DataPoint `json:"security_score_trend"`
	WarningsTrend      []DataPoint `json:"warnings_trend"`
	TestsTrend         []DataPoint `json:"tests_trend"`
	Period             string      `json:"period"` // "7d", "30d", "90d"
}

// DataPoint represents a single data point in time series
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// HistoryManager manages audit history
type HistoryManager struct {
	config HistoryConfig
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(storagePath string) *HistoryManager {
	if storagePath == "" {
		storagePath = "./history"
	}

	// Create storage directory if it doesn't exist
	os.MkdirAll(storagePath, 0755)

	return &HistoryManager{
		config: HistoryConfig{
			StoragePath:       storagePath,
			MaxFullRecords:    90,                   // Keep 90 days of full data
			RetentionDays:     365,                  // Keep 1 year of summary data
			CompressionAge:    30 * 24 * time.Hour, // Compress data older than 30 days
			EnableCompression: true,
		},
	}
}

// SaveAudit saves the current audit data to history
func (hm *HistoryManager) SaveAudit(data map[string]string, compliance ComplianceAnalysis) error {
	record := AuditRecord{
		Timestamp:      time.Now(),
		HardeningIndex: data["hardening_index"],
		Warnings:       data["warnings"],
		TestsPerformed: data["lynis_tests_done"],
		Suggestions:    countSuggestionsFromData(data),
		ComplianceScores: map[string]float64{
			"cis_level1": compliance.CIS_Level1.Score,
			"cis_level2": compliance.CIS_Level2.Score,
			"iso27001":   compliance.ISO27001.Score,
			"nist":       compliance.NIST.Score,
			"pcidss":     compliance.PCIDSS.Score,
		},
		KeyMetrics: extractKeyMetrics(data),
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("audit_%s.json", record.Timestamp.Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(hm.config.StoragePath, filename)

	// Save to file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(record); err != nil {
		return fmt.Errorf("failed to encode history data: %w", err)
	}

	// Run cleanup in background
	go hm.CleanupOldRecords()

	return nil
}

// GetTrend returns trend data for specified period
func (hm *HistoryManager) GetTrend(period string) (*TrendData, error) {
	var duration time.Duration
	switch period {
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	case "90d":
		duration = 90 * 24 * time.Hour
	default:
		duration = 30 * 24 * time.Hour
	}

	records, err := hm.GetRecordsSince(time.Now().Add(-duration))
	if err != nil {
		return nil, err
	}

	trend := &TrendData{
		Period:             period,
		SecurityScoreTrend: []DataPoint{},
		WarningsTrend:      []DataPoint{},
		TestsTrend:         []DataPoint{},
	}

	for _, record := range records {
		// Parse values
		score := parseFloat(record.HardeningIndex)
		warnings := parseFloat(record.Warnings)
		tests := parseFloat(record.TestsPerformed)

		trend.SecurityScoreTrend = append(trend.SecurityScoreTrend, DataPoint{
			Timestamp: record.Timestamp,
			Value:     score,
		})

		trend.WarningsTrend = append(trend.WarningsTrend, DataPoint{
			Timestamp: record.Timestamp,
			Value:     warnings,
		})

		trend.TestsTrend = append(trend.TestsTrend, DataPoint{
			Timestamp: record.Timestamp,
			Value:     tests,
		})
	}

	return trend, nil
}

// GetRecordsSince returns all records since specified time
func (hm *HistoryManager) GetRecordsSince(since time.Time) ([]AuditRecord, error) {
	files, err := os.ReadDir(hm.config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var records []AuditRecord

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(hm.config.StoragePath, file.Name())
		
		// Try to read record
		record, err := hm.readRecord(filePath)
		if err != nil {
			continue // Skip corrupted files
		}

		// Filter by time
		if record.Timestamp.After(since) {
			records = append(records, *record)
		}
	}

	// Sort by timestamp
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.Before(records[j].Timestamp)
	})

	return records, nil
}

// GetLatestRecord returns the most recent audit record
func (hm *HistoryManager) GetLatestRecord() (*AuditRecord, error) {
	files, err := os.ReadDir(hm.config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no history records found")
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := files[i].Info()
		infoJ, _ := files[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Read the latest file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(hm.config.StoragePath, file.Name())
		record, err := hm.readRecord(filePath)
		if err == nil {
			return record, nil
		}
	}

	return nil, fmt.Errorf("no valid records found")
}

// CompareWithPrevious compares current audit with previous one
func (hm *HistoryManager) CompareWithPrevious(current map[string]string) (map[string]interface{}, error) {
	previous, err := hm.GetLatestRecord()
	if err != nil {
		return nil, err
	}

	currentScore := parseFloat(current["hardening_index"])
	previousScore := parseFloat(previous.HardeningIndex)
	
	currentWarnings := parseFloat(current["warnings"])
	previousWarnings := parseFloat(previous.Warnings)

	comparison := map[string]interface{}{
		"score_change":    currentScore - previousScore,
		"warnings_change": currentWarnings - previousWarnings,
		"improved":        currentScore > previousScore,
		"previous_date":   previous.Timestamp,
		"days_since":      time.Since(previous.Timestamp).Hours() / 24,
	}

	return comparison, nil
}

// CleanupOldRecords removes old records based on retention policy
func (hm *HistoryManager) CleanupOldRecords() error {
	files, err := os.ReadDir(hm.config.StoragePath)
	if err != nil {
		return err
	}

	cutoffDate := time.Now().AddDate(0, 0, -hm.config.RetentionDays)
	compressionDate := time.Now().Add(-hm.config.CompressionAge)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(hm.config.StoragePath, file.Name())
		record, err := hm.readRecord(filePath)
		if err != nil {
			continue
		}

		// Delete very old records
		if record.Timestamp.Before(cutoffDate) {
			os.Remove(filePath)
			continue
		}

		// Compress old records (if enabled and not already compressed)
		if hm.config.EnableCompression && !record.Compressed && record.Timestamp.Before(compressionDate) {
			hm.compressRecord(filePath)
		}
	}

	return nil
}

// GetStorageStats returns storage usage statistics
func (hm *HistoryManager) GetStorageStats() (map[string]interface{}, error) {
	files, err := os.ReadDir(hm.config.StoragePath)
	if err != nil {
		return nil, err
	}

	var totalSize int64
	var compressedCount int
	var recordCount int

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
		recordCount++

		if filepath.Ext(file.Name()) == ".gz" {
			compressedCount++
		}
	}

	stats := map[string]interface{}{
		"total_records":     recordCount,
		"compressed_count":  compressedCount,
		"total_size_bytes":  totalSize,
		"total_size_kb":     totalSize / 1024,
		"total_size_mb":     float64(totalSize) / (1024 * 1024),
		"avg_record_size":   totalSize / int64(max(recordCount, 1)),
		"storage_path":      hm.config.StoragePath,
		"compression_ratio": float64(compressedCount) / float64(max(recordCount, 1)) * 100,
	}

	return stats, nil
}

// Helper functions

func (hm *HistoryManager) readRecord(filePath string) (*AuditRecord, error) {
	var reader io.Reader
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check if file is gzipped
	if filepath.Ext(filePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = file
	}

	var record AuditRecord
	if err := json.NewDecoder(reader).Decode(&record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (hm *HistoryManager) compressRecord(filePath string) error {
	// Read original file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Create compressed file
	gzPath := filePath + ".gz"
	gzFile, err := os.Create(gzPath)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	_, err = gzWriter.Write(data)
	if err != nil {
		return err
	}

	// Remove original file after successful compression
	return os.Remove(filePath)
}

func extractKeyMetrics(data map[string]string) map[string]string {
	// Extract only the most important metrics to save space
	metrics := make(map[string]string)
	
	importantKeys := []string{
		"os", "hostname", "kernel_version", "firewall_software",
		"firewall_status", "ssh_daemon_status", "logging_daemon",
	}

	for _, key := range importantKeys {
		if val, exists := data[key]; exists {
			metrics[key] = val
		}
	}

	return metrics
}

func countSuggestionsFromData(data map[string]string) int {
	count := 0
	for key, value := range data {
		if key == "suggestion" || key == "suggestion[]" {
			// Count pipe-separated suggestions
			count += len(splitByPipe(value))
		}
	}
	return max(count, 1)
}

func splitByPipe(s string) []string {
	if s == "" {
		return []string{}
	}
	return []string{s} // Simplified for now
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

