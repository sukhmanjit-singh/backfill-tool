package internal

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type PostmanCollection struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item []PostmanItem `json:"item"`
}

type PostmanItem struct {
	Name    string         `json:"name"`
	Request PostmanRequest `json:"request"`
	Item    []PostmanItem  `json:"item"` // For nested folders
}

type PostmanRequest struct {
	Method string          `json:"method"`
	URL    PostmanURL      `json:"url"` // Can be string or object
	Header []PostmanHeader `json:"header"`
	Body   PostmanBody     `json:"body"`
}

type PostmanURL struct {
	Raw string `json:"raw"`
}

type PostmanBody struct {
	Raw string `json:"raw"`
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func RunBatch(batchSize int, threads int, collection string, csv string) {
	jsonFile, err := os.Open(collection)
	if err != nil {
		fmt.Println("Error opening collection file:", err)
		return
	}
	defer jsonFile.Close()

	var postmanCollection PostmanCollection
	json.NewDecoder(jsonFile).Decode(&postmanCollection)

	fmt.Println("Collection name:", postmanCollection.Info.Name)
	fmt.Println("Number of items:", len(postmanCollection.Item))
	for _, item := range postmanCollection.Item {
		runItem(item)
		requestList, _ := ReadCSV(csv)
		recordsChan := make(chan map[string]string, len(requestList))
		resultsChan := make(chan string, len(requestList))
		numWorkers := 3
		var wg sync.WaitGroup
		for i := 1; i <= numWorkers; i++ {
			wg.Add(1)
			go worker(i, item, recordsChan, resultsChan, &wg)
		}
		for _, record := range requestList {
			recordsChan <- record
		}
		close(recordsChan) // No more records

		// Wait for all workers to finish
		wg.Wait()
		close(resultsChan) // No more results

		// Collect results
		fmt.Println("\n--- Results ---")
		for result := range resultsChan {
			fmt.Println(result)
		}
	}

}

func worker(id int, item PostmanItem, records chan map[string]string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for csvRow := range records {
		fmt.Printf("Worker %d processing record)\n", id)
		csvData := make(map[string]interface{})
		for column, value := range csvRow {
			csvData[column] = value
		}
		ReplaceJSONValues(item.Request.Body.Raw, csvData)
		modifiedBody, err := ReplaceJSONValues(item.Request.Body.Raw, csvData)
		if err != nil {
			// handle error
		}
		fmt.Println("Req Body:", modifiedBody)
		req, err := http.NewRequest(item.Request.Method, item.Request.URL.Raw, bytes.NewBufferString(modifiedBody))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		for _, header := range item.Request.Header {
			if header.Key == "" {
				continue
			}
			if header.Value == "" {
				continue
			}
			req.Header.Set(header.Key, header.Value)
		}
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		// Use modifiedBody in your HTTP request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		// Read response
		respB, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}
		fmt.Println("Status Code:", resp.StatusCode)
		results <- fmt.Sprint("Response:", string(respB))
	}
}

func ReadCSV(filepath string) ([]map[string]string, error) {
	// Open the file
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

	// Check if file is empty
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// First row is headers
	headers := records[0]

	// Convert each row to a map
	var rows []map[string]string
	for i := 1; i < len(records); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = records[i][j]
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ReplaceJSONValues replaces values in a JSON string with values from a map
func ReplaceJSONValues(jsonString string, replacements map[string]interface{}) (string, error) {
	// Parse JSON string into a map
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	// Replace values recursively
	replaceValues(jsonData, replacements)

	// Convert back to JSON string
	modifiedJSON, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("error converting to JSON: %v", err)
	}

	return string(modifiedJSON), nil
}

// replaceValues recursively replaces values in nested JSON
func replaceValues(data map[string]interface{}, replacements map[string]interface{}) {
	for key, value := range data {
		// Check if this key exists in replacements
		if newValue, exists := replacements[key]; exists {
			data[key] = newValue
		} else {
			// If value is a nested object, recurse into it
			switch v := value.(type) {
			case map[string]interface{}:
				replaceValues(v, replacements)
			case []interface{}:
				// Handle arrays of objects
				for _, item := range v {
					if obj, ok := item.(map[string]interface{}); ok {
						replaceValues(obj, replacements)
					}
				}
			}
		}
	}
}

func runItem(item PostmanItem) {
	fmt.Println("Item name:", item.Name)
	fmt.Println("Item request method:", item.Request.Method)
	fmt.Println("Item request URL:", item.Request.URL)
	fmt.Println("Item request header:", item.Request.Header)
	fmt.Println("Item request body:", item.Request.Body)
	fmt.Println("Item request URL raw:", item.Request.URL.Raw)
}
