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
	"time"
)

// PostmanCollection represents the top-level structure of a Postman collection JSON file
type PostmanCollection struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanItem represents a single request or folder in the Postman collection
// Items can be nested to represent folders containing other items
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
}

// PostmanURL represents the URL structure in Postman collections
// Postman can represent URLs as objects with multiple fields, but we primarily use the raw string
type PostmanURL struct {
	Raw   string        `json:"raw"`
	Query []QueryParam  `json:"query,omitempty"`
}

// QueryParam represents a query parameter in the URL
type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody represents the request body in a Postman request
type PostmanBody struct {
	Mode string `json:"mode,omitempty"` // raw, urlencoded, formdata, etc.
	Raw  string `json:"raw,omitempty"`
}

// PostmanHeader represents an HTTP header key-value pair
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunBatch is the main entry point for processing a Postman collection with CSV data
// It orchestrates the entire backfill process:
// 1. Loads the Postman collection
// 2. Reads the CSV data once
// 3. Processes all requests (including nested folders) with the specified number of worker threads
//
// Parameters:
//   - batchSize: Number of records to process per batch (currently informational)
//   - threads: Number of concurrent worker goroutines to spawn
//   - collection: Path to the Postman collection JSON file
//   - csv: Path to the CSV file containing data to backfill
func RunBatch(batchSize int, threads int, collection string, csv string) {
	// Validate input parameters
	if collection == "" {
		fmt.Println("Error: Collection file path is required")
		return
	}
	if csv == "" {
		fmt.Println("Error: CSV file path is required")
		return
	}
	if threads <= 0 {
		fmt.Println("Error: Number of threads must be greater than 0")
		return
	}

	// Load and parse the Postman collection
	jsonFile, err := os.Open(collection)
	if err != nil {
		fmt.Printf("Error opening collection file '%s': %v\n", collection, err)
		return
	}
	defer jsonFile.Close()

	var postmanCollection PostmanCollection
	if err := json.NewDecoder(jsonFile).Decode(&postmanCollection); err != nil {
		fmt.Printf("Error parsing collection JSON: %v\n", err)
		return
	}

	fmt.Printf("📦 Collection: %s\n", postmanCollection.Info.Name)
	fmt.Printf("📊 Items found: %d\n", len(postmanCollection.Item))

	// Read CSV data once and reuse for all requests (optimization)
	fmt.Printf("📂 Reading CSV file: %s\n", csv)
	requestList, err := ReadCSV(csv)
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		return
	}
	fmt.Printf("✓ Loaded %d records from CSV\n\n", len(requestList))

	if len(requestList) == 0 {
		fmt.Println("Warning: No data records found in CSV file (only headers)")
		return
	}

	// Process all items in the collection recursively (including nested folders)
	for _, item := range postmanCollection.Item {
		processItem(item, requestList, threads, 0)
	}

	fmt.Println("\n✓ All requests completed")
}

// processItem recursively processes a Postman item (request or folder)
// If the item has nested items (folder), it processes them recursively
// If the item is a request, it executes it with all CSV records using worker threads
//
// Parameters:
//   - item: The Postman item to process
//   - requestList: CSV data as a slice of maps (each map is one CSV row)
//   - threads: Number of concurrent workers to use
//   - depth: Current nesting depth (for formatting output)
func processItem(item PostmanItem, requestList []map[string]string, threads int, depth int) {
	indent := strings.Repeat("  ", depth)

	// Check if this is a folder (has nested items)
	if len(item.Item) > 0 {
		fmt.Printf("%s📁 Folder: %s\n", indent, item.Name)
		// Recursively process all items in the folder
		for _, nestedItem := range item.Item {
			processItem(nestedItem, requestList, threads, depth+1)
		}
		return
	}

	// This is a request item (not a folder)
	fmt.Printf("%s🔧 Processing request: %s\n", indent, item.Name)
	fmt.Printf("%s   Method: %s\n", indent, item.Request.Method)
	fmt.Printf("%s   URL: %s\n", indent, item.Request.URL.Raw)
	fmt.Printf("%s   Records to process: %d\n", indent, len(requestList))
	fmt.Printf("%s   Workers: %d\n", indent, threads)

	// Create channels for distributing work and collecting results
	recordsChan := make(chan map[string]string, len(requestList))
	resultsChan := make(chan RequestResult, len(requestList))

	// Create a wait group to synchronize worker goroutines
	var wg sync.WaitGroup

	// Spawn worker goroutines
	for i := 1; i <= threads; i++ {
		wg.Add(1)
		go worker(i, item, recordsChan, resultsChan, &wg, indent)
	}

	// Distribute all CSV records to workers through the channel
	for _, record := range requestList {
		recordsChan <- record
	}
	close(recordsChan) // Signal no more records will be sent

	// Wait for all workers to complete
	wg.Wait()
	close(resultsChan) // Signal no more results will be sent

	// Collect and display results
	successCount := 0
	failureCount := 0
	fmt.Printf("%s\n%s📊 Results:\n", indent, indent)
	for result := range resultsChan {
		if result.Success {
			successCount++
			fmt.Printf("%s   ✓ [%d] %s - %s\n", indent, result.StatusCode, result.Message, result.RecordInfo)
		} else {
			failureCount++
			fmt.Printf("%s   ✗ [ERROR] %s - %s\n", indent, result.Message, result.RecordInfo)
		}
	}

	fmt.Printf("%s\n%s✓ Completed: %d successful, %d failed\n\n", indent, indent, successCount, failureCount)
}

// RequestResult represents the outcome of a single HTTP request
type RequestResult struct {
	Success    bool
	StatusCode int
	Message    string
	RecordInfo string // Brief info about which CSV record was processed
}

// worker is a goroutine that processes CSV records and executes HTTP requests
// Multiple workers run concurrently to achieve parallelism
//
// Parameters:
//   - id: Worker identifier for logging
//   - item: The Postman request item to execute
//   - records: Channel from which to receive CSV records
//   - results: Channel to send request results
//   - wg: WaitGroup for synchronization
//   - indent: Indentation string for formatted output
func worker(id int, item PostmanItem, records chan map[string]string, results chan RequestResult, wg *sync.WaitGroup, indent string) {
	defer wg.Done()

	// Process records until the channel is closed
	for csvRow := range records {
		// Convert CSV row to a format suitable for template replacement
		csvData := make(map[string]interface{})
		for column, value := range csvRow {
			csvData[column] = value
		}

		// Generate a brief identifier for this record (for logging)
		recordInfo := getRecordInfo(csvRow)

		// Replace variables in the request URL (path variables and query parameters)
		finalURL, err := replaceURLVariables(item.Request.URL.Raw, csvRow)
		if err != nil {
			results <- RequestResult{
				Success:    false,
				Message:    fmt.Sprintf("Error processing URL: %v", err),
				RecordInfo: recordInfo,
			}
			continue
		}

		// Replace variables in the request body
		var modifiedBody string
		if item.Request.Body.Raw != "" {
			modifiedBody, err = ReplaceJSONValues(item.Request.Body.Raw, csvData)
			if err != nil {
				// If JSON parsing fails, treat body as plain text and replace template variables
				modifiedBody = replaceTemplateVariables(item.Request.Body.Raw, csvRow)
			}
		}

		// Create the HTTP request
		req, err := http.NewRequest(item.Request.Method, finalURL, bytes.NewBufferString(modifiedBody))
		if err != nil {
			results <- RequestResult{
				Success:    false,
				Message:    fmt.Sprintf("Error creating request: %v", err),
				RecordInfo: recordInfo,
			}
			continue
		}

		// Set headers with variable replacement
		for _, header := range item.Request.Header {
			if header.Key == "" || header.Value == "" {
				continue
			}
			// Replace variables in header values
			headerValue := replaceTemplateVariables(header.Value, csvRow)
			req.Header.Set(header.Key, headerValue)
		}

		// Set default Content-Type for JSON bodies if not already set
		if modifiedBody != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Execute the HTTP request with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			results <- RequestResult{
				Success:    false,
				Message:    fmt.Sprintf("Request failed: %v", err),
				RecordInfo: recordInfo,
			}
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			results <- RequestResult{
				Success:    false,
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("Error reading response: %v", err),
				RecordInfo: recordInfo,
			}
			continue
		}

		// Determine if request was successful based on status code
		success := resp.StatusCode >= 200 && resp.StatusCode < 300
		message := string(respBody)
		if len(message) > 100 {
			message = message[:100] + "..."
		}

		results <- RequestResult{
			Success:    success,
			StatusCode: resp.StatusCode,
			Message:    message,
			RecordInfo: recordInfo,
		}
	}
}

// getRecordInfo creates a brief string representation of a CSV record for logging
// It uses the first few fields to identify the record
func getRecordInfo(record map[string]string) string {
	if len(record) == 0 {
		return "empty record"
	}

	// Try to use common identifier fields
	for _, key := range []string{"id", "ID", "name", "Name", "email", "Email"} {
		if val, ok := record[key]; ok && val != "" {
			return fmt.Sprintf("%s=%s", key, val)
		}
	}

	// Otherwise, use the first field
	for key, val := range record {
		return fmt.Sprintf("%s=%s", key, val)
	}

	return "record"
}

// replaceURLVariables replaces template variables in the URL
// It handles both path variables (e.g., /users/{{userId}}) and query parameters
//
// Template syntax: {{variableName}} gets replaced with the value from CSV
//
// Examples:
//   - /api/users/{{userId}} -> /api/users/123
//   - /api/search?q={{query}}&limit={{limit}} -> /api/search?q=test&limit=10
func replaceURLVariables(rawURL string, csvData map[string]string) (string, error) {
	// Replace template variables in the URL
	finalURL := replaceTemplateVariables(rawURL, csvData)

	// Validate the URL
	_, err := url.Parse(finalURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL after replacement: %v", err)
	}

	return finalURL, nil
}

// replaceTemplateVariables replaces all {{variableName}} patterns in a string
// with corresponding values from the CSV data
//
// Parameters:
//   - template: String containing {{variable}} patterns
//   - data: Map of variable names to values
//
// Returns: String with all variables replaced
func replaceTemplateVariables(template string, data map[string]string) string {
	// Regular expression to match {{variableName}} patterns
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	// Replace all matches
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.TrimSpace(match[2 : len(match)-2])

		// Look up value in CSV data
		if value, exists := data[varName]; exists {
			return value
		}

		// If variable not found in CSV, keep the original placeholder
		return match
	})

	return result
}

// ReadCSV reads a CSV file and returns its contents as a slice of maps
// Each map represents one row, with column headers as keys
//
// Parameters:
//   - filepath: Path to the CSV file
//
// Returns:
//   - Slice of maps, where each map represents a CSV row
//   - Error if file cannot be read or parsed
func ReadCSV(filepath string) ([]map[string]string, error) {
	// Open the CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	// Validate CSV has content
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// First row contains headers
	headers := records[0]

	// Validate headers
	if len(headers) == 0 {
		return nil, fmt.Errorf("CSV file has no headers")
	}

	// Convert each data row to a map
	var rows []map[string]string
	for i := 1; i < len(records); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = records[i][j]
			} else {
				// Handle rows with fewer columns than headers
				row[header] = ""
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ReplaceJSONValues replaces values in a JSON string with values from CSV data
// It handles nested JSON objects and arrays
//
// This function uses two strategies:
// 1. Direct key matching: If a JSON key matches a CSV column, replace the value
// 2. Template matching: Replace {{variableName}} patterns in string values
//
// Parameters:
//   - jsonString: JSON string (typically from request body)
//   - replacements: Map of CSV column names to values
//
// Returns:
//   - Modified JSON string with values replaced
//   - Error if JSON is invalid
func ReplaceJSONValues(jsonString string, replacements map[string]interface{}) (string, error) {
	// Handle empty JSON
	if strings.TrimSpace(jsonString) == "" {
		return jsonString, nil
	}

	// Parse JSON string into a map or array
	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	// Replace values recursively
	replaceValuesRecursive(jsonData, replacements)

	// Convert back to JSON string
	modifiedJSON, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("error converting to JSON: %v", err)
	}

	return string(modifiedJSON), nil
}

// replaceValuesRecursive recursively processes JSON data structures and replaces values
// It handles maps (objects), slices (arrays), and string template patterns
func replaceValuesRecursive(data interface{}, replacements map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair in the object
		for key, value := range v {
			// Strategy 1: Direct key matching
			if newValue, exists := replacements[key]; exists {
				v[key] = newValue
			} else {
				// Strategy 2: If value is a string with templates, replace them
				if strValue, ok := value.(string); ok {
					// Convert replacements to string map for template replacement
					strReplacements := make(map[string]string)
					for k, val := range replacements {
						strReplacements[k] = fmt.Sprintf("%v", val)
					}
					v[key] = replaceTemplateVariables(strValue, strReplacements)
				} else {
					// Recurse into nested structures
					replaceValuesRecursive(value, replacements)
				}
			}
		}

	case []interface{}:
		// Process each item in the array
		for i, item := range v {
			if strValue, ok := item.(string); ok {
				// Replace templates in string array items
				strReplacements := make(map[string]string)
				for k, val := range replacements {
					strReplacements[k] = fmt.Sprintf("%v", val)
				}
				v[i] = replaceTemplateVariables(strValue, strReplacements)
			} else {
				// Recurse into nested structures
				replaceValuesRecursive(item, replacements)
			}
		}
	}
}
