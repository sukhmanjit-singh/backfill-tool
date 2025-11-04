package internal

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// RunConfig contains all configuration for a batch run
type RunConfig struct {
	BatchSize    int
	Threads      int
	Collection   string
	CSV          string
	MetricsFile  string
	Verbose      bool
	Quiet        bool
	BearerToken  string // CLI override for bearer token
}

// PostmanCollection represents the top-level structure of a Postman collection JSON file
type PostmanCollection struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item []PostmanItem `json:"item"`
	Auth *PostmanAuth  `json:"auth,omitempty"` // Collection-level auth
}

// PostmanItem represents a single request or folder in the Postman collection
type PostmanItem struct {
	Name    string         `json:"name"`
	Request PostmanRequest `json:"request"`
	Item    []PostmanItem  `json:"item"` // For nested folders
}

// PostmanRequest contains all the details needed to execute an HTTP request
type PostmanRequest struct {
	Method string          `json:"method"`
	URL    PostmanURL      `json:"url"`
	Header []PostmanHeader `json:"header"`
	Body   PostmanBody     `json:"body"`
	Auth   *PostmanAuth    `json:"auth,omitempty"` // Request-level auth
}

// PostmanURL represents the URL structure in Postman collections
type PostmanURL struct {
	Raw   string       `json:"raw"`
	Query []QueryParam `json:"query,omitempty"`
}

// QueryParam represents a query parameter in the URL
type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody represents the request body in a Postman request
type PostmanBody struct {
	Mode string `json:"mode,omitempty"`
	Raw  string `json:"raw,omitempty"`
}

// PostmanHeader represents an HTTP header key-value pair
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanAuth represents authentication configuration in Postman
type PostmanAuth struct {
	Type   string          `json:"type"` // "bearer", "apikey", "basic", etc.
	Bearer []PostmanKV     `json:"bearer,omitempty"`
	APIKey []PostmanKV     `json:"apikey,omitempty"`
	Basic  []PostmanKV     `json:"basic,omitempty"`
}

// PostmanKV represents key-value pairs in auth configuration
type PostmanKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// RequestResult represents the outcome of a single HTTP request
type RequestResult struct {
	Success       bool
	StatusCode    int
	ResponseTime  time.Duration
	Message       string
	RecordInfo    string
	Error         string
	URL           string
	Method        string
	CSVData       map[string]string
	RequestName   string
	Timestamp     time.Time
}

// RequestMetrics tracks statistics for a request or collection item
type RequestMetrics struct {
	Name           string
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	TotalTime      time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	StartTime      time.Time
	EndTime        time.Time
	FailedRequests []RequestResult
}

// RunMetrics tracks overall execution metrics
type RunMetrics struct {
	CollectionName string
	CSVFile        string
	StartTime      time.Time
	EndTime        time.Time
	TotalRecords   int
	ItemMetrics    []RequestMetrics
}

// ProgressTracker manages real-time progress display
type ProgressTracker struct {
	total       int64
	current     int64
	success     int64
	failure     int64
	startTime   time.Time
	quiet       bool
	mu          sync.Mutex
	lastPrint   time.Time
	description string
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int, description string, quiet bool) *ProgressTracker {
	return &ProgressTracker{
		total:       int64(total),
		current:     0,
		success:     0,
		failure:     0,
		startTime:   time.Now(),
		quiet:       quiet,
		lastPrint:   time.Now(),
		description: description,
	}
}

// Update increments progress and updates display
func (p *ProgressTracker) Update(success bool) {
	atomic.AddInt64(&p.current, 1)
	if success {
		atomic.AddInt64(&p.success, 1)
	} else {
		atomic.AddInt64(&p.failure, 1)
	}

	if !p.quiet {
		p.mu.Lock()
		// Update display every 100ms to avoid flickering
		if time.Since(p.lastPrint) > 100*time.Millisecond {
			p.display()
			p.lastPrint = time.Now()
		}
		p.mu.Unlock()
	}
}

// Finish completes the progress display
func (p *ProgressTracker) Finish() {
	if !p.quiet {
		p.mu.Lock()
		p.display()
		fmt.Println() // New line after progress
		p.mu.Unlock()
	}
}

// display renders the progress bar
func (p *ProgressTracker) display() {
	current := atomic.LoadInt64(&p.current)
	success := atomic.LoadInt64(&p.success)
	failure := atomic.LoadInt64(&p.failure)

	percent := float64(current) / float64(p.total) * 100
	elapsed := time.Since(p.startTime)
	avgTime := elapsed / time.Duration(current+1)
	eta := avgTime * time.Duration(p.total-current)

	// Create progress bar (40 characters wide)
	barWidth := 40
	filled := int(float64(barWidth) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Format output with colors
	fmt.Printf("\r%sProgress:%s [%s] %d/%d (%.1f%%) | %s✓%d%s %s✗%d%s | Avg: %dms | ETA: %s  ",
		colorBold, colorReset,
		bar,
		current, p.total, percent,
		colorGreen, success, colorReset,
		colorRed, failure, colorReset,
		avgTime.Milliseconds(),
		formatDuration(eta))
}

// formatDuration formats duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// colorize applies color to text
func colorize(color, text string) string {
	return color + text + colorReset
}

// RunBatch is the main entry point for processing a Postman collection with CSV data
func RunBatch(config RunConfig) {
	startTime := time.Now()

	// Validate input parameters
	if config.Collection == "" {
		fmt.Println(colorize(colorRed, "Error: Collection file path is required"))
		return
	}
	if config.CSV == "" {
		fmt.Println(colorize(colorRed, "Error: CSV file path is required"))
		return
	}
	if config.Threads <= 0 {
		fmt.Println(colorize(colorRed, "Error: Number of threads must be greater than 0"))
		return
	}

	// Load and parse the Postman collection
	jsonFile, err := os.Open(config.Collection)
	if err != nil {
		fmt.Printf("%s\n", colorize(colorRed, fmt.Sprintf("Error opening collection file '%s': %v", config.Collection, err)))
		return
	}
	defer jsonFile.Close()

	var postmanCollection PostmanCollection
	if err := json.NewDecoder(jsonFile).Decode(&postmanCollection); err != nil {
		fmt.Printf("%s\n", colorize(colorRed, fmt.Sprintf("Error parsing collection JSON: %v", err)))
		return
	}

	if !config.Quiet {
		fmt.Printf("%s\n", colorize(colorCyan+colorBold, "📦 Collection: "+postmanCollection.Info.Name))
		fmt.Printf("📊 Items found: %s\n", colorize(colorYellow, fmt.Sprintf("%d", len(postmanCollection.Item))))
	}

	// Read CSV data once and reuse for all requests
	if !config.Quiet {
		fmt.Printf("📂 Reading CSV file: %s\n", config.CSV)
	}
	requestList, err := ReadCSV(config.CSV)
	if err != nil {
		fmt.Printf("%s\n", colorize(colorRed, fmt.Sprintf("Error reading CSV file: %v", err)))
		return
	}

	if !config.Quiet {
		fmt.Printf("%s\n\n", colorize(colorGreen, fmt.Sprintf("✓ Loaded %d records from CSV", len(requestList))))
	}

	if len(requestList) == 0 {
		fmt.Println(colorize(colorYellow, "Warning: No data records found in CSV file (only headers)"))
		return
	}

	// Initialize run metrics
	runMetrics := &RunMetrics{
		CollectionName: postmanCollection.Info.Name,
		CSVFile:        config.CSV,
		StartTime:      startTime,
		TotalRecords:   len(requestList),
		ItemMetrics:    []RequestMetrics{},
	}

	// Process all items in the collection recursively
	for _, item := range postmanCollection.Item {
		processItem(item, requestList, config, runMetrics, 0, postmanCollection.Auth)
	}

	runMetrics.EndTime = time.Now()

	// Save metrics to file
	if err := saveMetrics(runMetrics, config); err != nil && config.Verbose {
		fmt.Printf("%s\n", colorize(colorYellow, fmt.Sprintf("Warning: Failed to save metrics: %v", err)))
	}

	// Print final summary
	if !config.Quiet {
		printFinalSummary(runMetrics)
	}
}

// processItem recursively processes a Postman item (request or folder)
func processItem(item PostmanItem, requestList []map[string]string, config RunConfig, runMetrics *RunMetrics, depth int, collectionAuth *PostmanAuth) {
	indent := strings.Repeat("  ", depth)

	// Check if this is a folder
	if len(item.Item) > 0 {
		if !config.Quiet {
			fmt.Printf("%s%s\n", indent, colorize(colorCyan, "📁 Folder: "+item.Name))
		}
		for _, nestedItem := range item.Item {
			processItem(nestedItem, requestList, config, runMetrics, depth+1, collectionAuth)
		}
		return
	}

	// This is a request item
	metrics := RequestMetrics{
		Name:           item.Name,
		TotalRequests:  int64(len(requestList)),
		SuccessCount:   0,
		FailureCount:   0,
		MinTime:        time.Hour, // Will be updated
		MaxTime:        0,
		StartTime:      time.Now(),
		FailedRequests: []RequestResult{},
	}

	if !config.Quiet {
		fmt.Printf("%s%s\n", indent, colorize(colorBold, "🔧 Processing: "+item.Name))
		fmt.Printf("%s   Method: %s | URL: %s\n", indent,
			colorize(colorPurple, item.Request.Method),
			colorize(colorGray, item.Request.URL.Raw))
		fmt.Printf("%s   Records: %s | Workers: %s\n", indent,
			colorize(colorYellow, fmt.Sprintf("%d", len(requestList))),
			colorize(colorYellow, fmt.Sprintf("%d", config.Threads)))
		fmt.Println()
	}

	// Create progress tracker
	progress := NewProgressTracker(len(requestList), item.Name, config.Quiet)

	// Create channels
	recordsChan := make(chan map[string]string, len(requestList))
	resultsChan := make(chan RequestResult, len(requestList))

	var wg sync.WaitGroup
	var mu sync.Mutex // Protect metrics updates

	// Spawn workers
	for i := 1; i <= config.Threads; i++ {
		wg.Add(1)
		go worker(i, item, recordsChan, resultsChan, &wg, config, collectionAuth)
	}

	// Distribute work
	for _, record := range requestList {
		recordsChan <- record
	}
	close(recordsChan)

	// Collect results in background
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Process results
	for result := range resultsChan {
		mu.Lock()
		if result.Success {
			metrics.SuccessCount++
		} else {
			metrics.FailureCount++
			metrics.FailedRequests = append(metrics.FailedRequests, result)
		}

		// Update timing metrics
		if result.ResponseTime < metrics.MinTime {
			metrics.MinTime = result.ResponseTime
		}
		if result.ResponseTime > metrics.MaxTime {
			metrics.MaxTime = result.ResponseTime
		}
		metrics.TotalTime += result.ResponseTime
		mu.Unlock()

		progress.Update(result.Success)
	}

	progress.Finish()
	metrics.EndTime = time.Now()

	// Save failed requests to CSV
	if len(metrics.FailedRequests) > 0 {
		failedFile := saveFailedRequests(metrics.FailedRequests, item.Name)
		if !config.Quiet && failedFile != "" {
			fmt.Printf("%s   %s\n", indent, colorize(colorYellow, fmt.Sprintf("❌ Failed: %d requests saved to %s", len(metrics.FailedRequests), failedFile)))
			fmt.Printf("%s   %s\n", indent, colorize(colorGray, "   (CSV includes error details: status code, message, URL, timestamp)"))
		}
	}

	// Print summary for this item
	if !config.Quiet {
		printRequestSummary(metrics, indent)
	}

	runMetrics.ItemMetrics = append(runMetrics.ItemMetrics, metrics)
}

// resolveAuth determines which auth to use based on hierarchy:
// 1. CLI override (--bearer-token flag)
// 2. Request-level auth
// 3. Collection-level auth
func resolveAuth(collectionAuth *PostmanAuth, requestAuth *PostmanAuth, cliToken string) *PostmanAuth {
	// CLI override takes precedence
	if cliToken != "" {
		return &PostmanAuth{
			Type: "bearer",
			Bearer: []PostmanKV{
				{Key: "token", Value: cliToken, Type: "string"},
			},
		}
	}

	// Request-level auth overrides collection auth
	if requestAuth != nil {
		return requestAuth
	}

	// Fall back to collection-level auth
	return collectionAuth
}

// applyAuth applies authentication to an HTTP request
// Supports bearer tokens, API keys, and basic auth with template variable replacement
func applyAuth(req *http.Request, auth *PostmanAuth, csvData map[string]string) {
	if auth == nil {
		return
	}

	switch auth.Type {
	case "bearer":
		// Extract bearer token value
		token := ""
		for _, kv := range auth.Bearer {
			if kv.Key == "token" {
				token = kv.Value
				break
			}
		}
		if token != "" {
			// Replace template variables in token
			token = replaceTemplateVariables(token, csvData)
			req.Header.Set("Authorization", "Bearer "+token)
		}

	case "apikey":
		// Extract API key header name and value
		keyName := ""
		keyValue := ""
		for _, kv := range auth.APIKey {
			if kv.Key == "key" {
				keyName = kv.Value
			} else if kv.Key == "value" {
				keyValue = kv.Value
			}
		}
		if keyName != "" && keyValue != "" {
			// Replace template variables in both key and value
			keyName = replaceTemplateVariables(keyName, csvData)
			keyValue = replaceTemplateVariables(keyValue, csvData)
			req.Header.Set(keyName, keyValue)
		}

	case "basic":
		// Extract username and password
		username := ""
		password := ""
		for _, kv := range auth.Basic {
			if kv.Key == "username" {
				username = kv.Value
			} else if kv.Key == "password" {
				password = kv.Value
			}
		}
		if username != "" {
			// Replace template variables
			username = replaceTemplateVariables(username, csvData)
			password = replaceTemplateVariables(password, csvData)
			req.SetBasicAuth(username, password)
		}
	}
}

// worker processes CSV records and executes HTTP requests
func worker(id int, item PostmanItem, records chan map[string]string, results chan RequestResult, wg *sync.WaitGroup, config RunConfig, collectionAuth *PostmanAuth) {
	defer wg.Done()

	for csvRow := range records {
		startTime := time.Now()

		csvData := make(map[string]interface{})
		for column, value := range csvRow {
			csvData[column] = value
		}

		recordInfo := getRecordInfo(csvRow)

		result := RequestResult{
			Timestamp:   startTime,
			RequestName: item.Name,
			Method:      item.Request.Method,
			CSVData:     csvRow,
			RecordInfo:  recordInfo,
		}

		// Replace URL variables (path variables and query parameters)
		finalURL, err := BuildURLWithQueryParams(item.Request.URL, csvRow)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Error processing URL: %v", err)
			result.ResponseTime = time.Since(startTime)
			results <- result
			continue
		}
		result.URL = finalURL

		// Replace body variables
		var modifiedBody string
		if item.Request.Body.Raw != "" {
			modifiedBody, err = ReplaceJSONValues(item.Request.Body.Raw, csvData)
			if err != nil {
				modifiedBody = replaceTemplateVariables(item.Request.Body.Raw, csvRow)
			}
		}

		// Create HTTP request
		req, err := http.NewRequest(item.Request.Method, finalURL, bytes.NewBufferString(modifiedBody))
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Error creating request: %v", err)
			result.ResponseTime = time.Since(startTime)
			results <- result
			continue
		}

		// Resolve and apply authentication
		auth := resolveAuth(collectionAuth, item.Request.Auth, config.BearerToken)
		applyAuth(req, auth, csvRow)

		// Set headers (after auth so explicit headers can override auth headers if needed)
		for _, header := range item.Request.Header {
			if header.Key == "" || header.Value == "" {
				continue
			}
			headerValue := replaceTemplateVariables(header.Value, csvRow)
			req.Header.Set(header.Key, headerValue)
		}

		// Default Content-Type
		if modifiedBody != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Execute request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Request failed: %v", err)
			result.ResponseTime = time.Since(startTime)
			results <- result
			continue
		}

		// Read response
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		result.ResponseTime = time.Since(startTime)
		result.StatusCode = resp.StatusCode
		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

		if err != nil {
			result.Error = fmt.Sprintf("Error reading response: %v", err)
			result.Success = false
		} else {
			message := string(respBody)
			if len(message) > 100 {
				message = message[:100] + "..."
			}
			result.Message = message

			if !result.Success {
				result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, message)
			}
		}

		results <- result
	}
}

// saveFailedRequests saves failed requests to a CSV file for retry
// The CSV includes original data columns PLUS error detail columns at the end
// This allows both: (1) easy retry by re-uploading, (2) viewing error details
// Error columns are ignored during retry since they don't match template variables
func saveFailedRequests(failedRequests []RequestResult, requestName string) string {
	if len(failedRequests) == 0 {
		return ""
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(requestName, " ", "_")
	filename := fmt.Sprintf("failed_requests_%s_%s.csv", safeName, timestamp)

	file, err := os.Create(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Collect all unique CSV column names from original data
	headers := []string{}
	headerMap := make(map[string]bool)
	for _, fr := range failedRequests {
		for key := range fr.CSVData {
			if !headerMap[key] {
				headers = append(headers, key)
				headerMap[key] = true
			}
		}
	}

	// Add error detail columns at the end (these will be ignored on retry)
	errorColumns := []string{
		"_error_status_code",
		"_error_message",
		"_error_url",
		"_error_method",
		"_error_timestamp",
		"_error_response_time_ms",
	}
	allHeaders := append(headers, errorColumns...)

	// Write header row
	writer.Write(allHeaders)

	// Write failed request data with error details
	for _, fr := range failedRequests {
		row := make([]string, len(allHeaders))

		// Fill original CSV columns
		for i, header := range headers {
			row[i] = fr.CSVData[header]
		}

		// Fill error detail columns
		offset := len(headers)
		row[offset+0] = fmt.Sprintf("%d", fr.StatusCode)
		row[offset+1] = cleanErrorMessage(fr.Error)
		row[offset+2] = fr.URL
		row[offset+3] = fr.Method
		row[offset+4] = fr.Timestamp.Format(time.RFC3339)
		row[offset+5] = fmt.Sprintf("%d", fr.ResponseTime.Milliseconds())

		writer.Write(row)
	}

	return filename
}

// cleanErrorMessage removes problematic characters from error messages for CSV
func cleanErrorMessage(errMsg string) string {
	// Replace newlines and carriage returns with spaces
	errMsg = strings.ReplaceAll(errMsg, "\n", " ")
	errMsg = strings.ReplaceAll(errMsg, "\r", " ")
	// Replace multiple spaces with single space
	errMsg = strings.Join(strings.Fields(errMsg), " ")
	// Truncate if too long
	if len(errMsg) > 500 {
		errMsg = errMsg[:497] + "..."
	}
	return errMsg
}

// saveMetrics saves execution metrics to JSON file
func saveMetrics(runMetrics *RunMetrics, config RunConfig) error {
	// Determine filename
	filename := config.MetricsFile
	if filename == "" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("metrics_%s.json", timestamp)
	}

	// Calculate summary statistics
	totalSuccess := int64(0)
	totalFailure := int64(0)
	totalRequests := int64(0)
	for _, item := range runMetrics.ItemMetrics {
		totalSuccess += item.SuccessCount
		totalFailure += item.FailureCount
		totalRequests += item.TotalRequests
	}

	// Create output structure
	output := map[string]interface{}{
		"collection_name": runMetrics.CollectionName,
		"csv_file":        runMetrics.CSVFile,
		"start_time":      runMetrics.StartTime.Format(time.RFC3339),
		"end_time":        runMetrics.EndTime.Format(time.RFC3339),
		"duration_seconds": runMetrics.EndTime.Sub(runMetrics.StartTime).Seconds(),
		"total_records":   runMetrics.TotalRecords,
		"summary": map[string]interface{}{
			"total_requests":   totalRequests,
			"successful":       totalSuccess,
			"failed":           totalFailure,
			"success_rate_pct": float64(totalSuccess) / float64(totalRequests) * 100,
		},
		"items": []map[string]interface{}{},
	}

	// Add per-item metrics
	items := []map[string]interface{}{}
	for _, item := range runMetrics.ItemMetrics {
		avgTime := time.Duration(0)
		if item.SuccessCount+item.FailureCount > 0 {
			avgTime = item.TotalTime / time.Duration(item.SuccessCount+item.FailureCount)
		}

		itemData := map[string]interface{}{
			"name":            item.Name,
			"total_requests":  item.TotalRequests,
			"successful":      item.SuccessCount,
			"failed":          item.FailureCount,
			"success_rate_pct": float64(item.SuccessCount) / float64(item.TotalRequests) * 100,
			"timing": map[string]interface{}{
				"avg_ms": avgTime.Milliseconds(),
				"min_ms": item.MinTime.Milliseconds(),
				"max_ms": item.MaxTime.Milliseconds(),
			},
			"duration_seconds": item.EndTime.Sub(item.StartTime).Seconds(),
		}
		items = append(items, itemData)
	}
	output["items"] = items

	// Write to file
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	if !config.Quiet {
		fmt.Printf("\n%s\n", colorize(colorGreen, "💾 Metrics saved to: "+filename))
	}

	return nil
}

// printRequestSummary prints summary for a single request
func printRequestSummary(metrics RequestMetrics, indent string) {
	fmt.Println()
	fmt.Printf("%s%s\n", indent, colorize(colorBold, "📊 Summary:"))

	successRate := float64(metrics.SuccessCount) / float64(metrics.TotalRequests) * 100
	avgTime := time.Duration(0)
	if metrics.SuccessCount+metrics.FailureCount > 0 {
		avgTime = metrics.TotalTime / time.Duration(metrics.SuccessCount+metrics.FailureCount)
	}

	fmt.Printf("%s   Total:        %s\n", indent, colorize(colorCyan, fmt.Sprintf("%d", metrics.TotalRequests)))
	fmt.Printf("%s   Successful:   %s (%.1f%%)\n", indent, colorize(colorGreen, fmt.Sprintf("%d", metrics.SuccessCount)), successRate)
	fmt.Printf("%s   Failed:       %s (%.1f%%)\n", indent, colorize(colorRed, fmt.Sprintf("%d", metrics.FailureCount)), 100-successRate)
	fmt.Printf("%s   Avg Time:     %dms\n", indent, avgTime.Milliseconds())
	fmt.Printf("%s   Min Time:     %dms\n", indent, metrics.MinTime.Milliseconds())
	fmt.Printf("%s   Max Time:     %dms\n", indent, metrics.MaxTime.Milliseconds())
	fmt.Printf("%s   Duration:     %s\n", indent, formatDuration(metrics.EndTime.Sub(metrics.StartTime)))
	fmt.Println()
}

// printFinalSummary prints overall execution summary
func printFinalSummary(runMetrics *RunMetrics) {
	totalSuccess := int64(0)
	totalFailure := int64(0)
	totalRequests := int64(0)

	for _, item := range runMetrics.ItemMetrics {
		totalSuccess += item.SuccessCount
		totalFailure += item.FailureCount
		totalRequests += item.TotalRequests
	}

	duration := runMetrics.EndTime.Sub(runMetrics.StartTime)
	throughput := float64(totalRequests) / duration.Seconds()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("%s\n", colorize(colorBold+colorCyan, "🎯 EXECUTION COMPLETE"))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Collection:     %s\n", runMetrics.CollectionName)
	fmt.Printf("Total Requests: %s\n", colorize(colorCyan, fmt.Sprintf("%d", totalRequests)))
	fmt.Printf("Successful:     %s (%.1f%%)\n", colorize(colorGreen, fmt.Sprintf("%d", totalSuccess)), float64(totalSuccess)/float64(totalRequests)*100)
	fmt.Printf("Failed:         %s (%.1f%%)\n", colorize(colorRed, fmt.Sprintf("%d", totalFailure)), float64(totalFailure)/float64(totalRequests)*100)
	fmt.Printf("Duration:       %s\n", colorize(colorYellow, formatDuration(duration)))
	fmt.Printf("Throughput:     %s req/s\n", colorize(colorYellow, fmt.Sprintf("%.2f", throughput)))
	fmt.Println(strings.Repeat("=", 60))
}

// getRecordInfo creates a brief string representation of a CSV record for logging
func getRecordInfo(record map[string]string) string {
	if len(record) == 0 {
		return "empty record"
	}

	for _, key := range []string{"id", "ID", "name", "Name", "email", "Email"} {
		if val, ok := record[key]; ok && val != "" {
			return fmt.Sprintf("%s=%s", key, val)
		}
	}

	for key, val := range record {
		return fmt.Sprintf("%s=%s", key, val)
	}

	return "record"
}

// replaceURLVariables replaces template variables in the URL
// Handles both path variables and query parameters with proper URL encoding
func replaceURLVariables(rawURL string, csvData map[string]string) (string, error) {
	// First, replace template variables in the raw URL
	finalURL := replaceTemplateVariables(rawURL, csvData)

	// Parse the URL to validate and potentially add query params
	parsedURL, err := url.Parse(finalURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL after replacement: %v", err)
	}

	return parsedURL.String(), nil
}

// BuildURLWithQueryParams constructs a complete URL with query parameters
// It handles both Postman's structured query params and raw URL query strings
// All query parameter values support template variable replacement
// Exported for testing purposes
func BuildURLWithQueryParams(postmanURL PostmanURL, csvData map[string]string) (string, error) {
	// Start with the raw URL and replace path variables
	baseURL := replaceTemplateVariables(postmanURL.Raw, csvData)

	// Parse the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %v", err)
	}

	// If Postman has structured query parameters, process them
	if len(postmanURL.Query) > 0 {
		queryParams := url.Values{}

		// First, preserve any existing query params from the raw URL
		existingParams := parsedURL.Query()
		for key, values := range existingParams {
			for _, value := range values {
				queryParams.Add(key, value)
			}
		}

		// Add/override with Postman's structured query parameters
		for _, param := range postmanURL.Query {
			if param.Key == "" {
				continue // Skip empty keys
			}

			// Replace template variables in the query parameter value
			paramValue := replaceTemplateVariables(param.Value, csvData)

			// Set the parameter (replaces existing values with same key)
			queryParams.Set(param.Key, paramValue)
		}

		// Build the final URL with encoded query parameters
		parsedURL.RawQuery = queryParams.Encode()
	}

	return parsedURL.String(), nil
}

// replaceTemplateVariables replaces all {{variableName}} patterns in a string
func replaceTemplateVariables(template string, data map[string]string) string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		varName := strings.TrimSpace(match[2 : len(match)-2])
		if value, exists := data[varName]; exists {
			return value
		}
		return match
	})
	return result
}

// ReadCSV reads a CSV file and returns its contents as a slice of maps
func ReadCSV(filepath string) ([]map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	if len(headers) == 0 {
		return nil, fmt.Errorf("CSV file has no headers")
	}

	var rows []map[string]string
	for i := 1; i < len(records); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = records[i][j]
			} else {
				row[header] = ""
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ReplaceJSONValues replaces values in a JSON string with values from CSV data
func ReplaceJSONValues(jsonString string, replacements map[string]interface{}) (string, error) {
	if strings.TrimSpace(jsonString) == "" {
		return jsonString, nil
	}

	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	replaceValuesRecursive(jsonData, replacements)

	modifiedJSON, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("error converting to JSON: %v", err)
	}

	return string(modifiedJSON), nil
}

// replaceValuesRecursive recursively processes JSON data structures and replaces values
func replaceValuesRecursive(data interface{}, replacements map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if newValue, exists := replacements[key]; exists {
				v[key] = newValue
			} else {
				if strValue, ok := value.(string); ok {
					strReplacements := make(map[string]string)
					for k, val := range replacements {
						strReplacements[k] = fmt.Sprintf("%v", val)
					}
					v[key] = replaceTemplateVariables(strValue, strReplacements)
				} else {
					replaceValuesRecursive(value, replacements)
				}
			}
		}

	case []interface{}:
		for i, item := range v {
			if strValue, ok := item.(string); ok {
				strReplacements := make(map[string]string)
				for k, val := range replacements {
					strReplacements[k] = fmt.Sprintf("%v", val)
				}
				v[i] = replaceTemplateVariables(strValue, strReplacements)
			} else {
				replaceValuesRecursive(item, replacements)
			}
		}
	}
}
