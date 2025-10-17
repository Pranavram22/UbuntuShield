package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	// Test data
	testData := map[string]string{
		"firewall_status":    "active",
		"ssh_daemon_status":  "running",
		"ssh_daemon_options": "PermitRootLogin yes",
		"logging_daemon":     "rsyslog",
	}

	fmt.Println("Testing individual compliance functions:")

	// Test SOC2
	soc2Result := analyzeSOC2(testData)
	fmt.Printf("SOC2: Score=%.1f, Total=%d, Passed=%d\n", soc2Result.Score, soc2Result.Total, soc2Result.Passed)

	// Test HIPAA
	hipaaResult := analyzeHIPAA(testData)
	fmt.Printf("HIPAA: Score=%.1f, Total=%d, Passed=%d\n", hipaaResult.Score, hipaaResult.Total, hipaaResult.Passed)

	// Test full analysis
	fullResult := analyzeCompliance(testData)

	// Convert to JSON to see what's actually being generated
	jsonData, err := json.MarshalIndent(fullResult, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nFull compliance analysis JSON:")
	fmt.Println(string(jsonData))
}
