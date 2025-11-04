# Backfill Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A high-performance CLI tool for executing bulk API requests using Postman collections and CSV data. Perfect for data migration, backfilling, bulk testing, and API automation tasks.

## 🚀 Features

## 🎯 New in v2.1.0

###  Real-time Progress Tracking
- **Live progress bars** showing completion percentage
- **Real-time metrics**: ✓ success count, ✗ failure count
- **Performance stats**: Average response time, ETA calculations
- **Colored output** for better visibility (green/red/yellow)

### 📊 Comprehensive Metrics & Analytics
- **Auto-save metrics** to JSON file after each run
- **Per-request statistics**: min/max/avg response times
- **Success rate tracking** with detailed breakdowns
- **Throughput measurements**: requests per second
- **Historical data**: Compare runs over time

### ⚡ Failed Request Management
- **Auto-save failed requests** to CSV (same format as input!)
- **One-command retry**: Use failed CSV directly as input
- **Separate files** for each request type
- **Perfect for debugging** and iterative testing

### 🛠️ Enhanced User Experience
- **Built-in examples**: Run `backfill-tool examples` for detailed use cases
- **Version info**: `backfill-tool version` shows features and build details
- **Improved help text**: Real-world scenarios and best practices
- **Template syntax guide** in CLI help

### 🔄 CI/CD Ready
- **Quiet mode** (`--quiet`): Suppress progress bars for clean logs
- **Metrics always saved**: Even in quiet mode
- **Exit codes**: Proper status codes for automation
- **JSON output**: Machine-readable execution summary


- **Postman Collection Support**: Import and execute Postman collections directly
- **CSV Data Integration**: Use CSV files to dynamically populate request parameters
- **Template Variable Replacement**: Support for `{{variable}}` syntax in:
  - URL path variables (e.g., `/api/users/{{userId}}`)
  - Query parameters (e.g., `?name={{name}}&year={{year}}`)
  - Request headers (e.g., `Authorization: Bearer {{token}}`)
  - Request bodies (JSON and text)
- **Nested Folder Support**: Recursively process nested folders in Postman collections
- **Concurrent Execution**: Configurable worker threads for parallel request processing
- **Multiple HTTP Methods**: Support for GET, POST, PUT, PATCH, DELETE, and all other HTTP methods
- **Comprehensive Error Handling**: Detailed error messages and status reporting
- **Production-Ready**: Clean, well-documented code with proper error handling

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Examples](#examples)
- [Template Variable Syntax](#template-variable-syntax)
- [Configuration Options](#configuration-options)
- [CSV File Format](#csv-file-format)
- [Postman Collection Setup](#postman-collection-setup)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## 🔧 Installation

### Prerequisites

- Go 1.25 or higher

### Build from Source

```bash
# Clone the repository
git clone https://github.com/sukhmanjit-singh/backfill-tool.git
cd backfill-tool

# Build the binary
go build -o backfill-tool .

# (Optional) Install globally
go install
```

### Using Go Install

```bash
go install github.com/sukhmanjit-singh/backfill-tool@latest
```

## ⚡ Quick Start

1. **Prepare your CSV data file** (e.g., `data.csv`):
```csv
name,year,userId
Alice,2020,1
Bob,2021,2
Charlie,2022,3
```

2. **Export your Postman collection** as JSON (e.g., `collection.json`)

3. **Run the tool**:
```bash
./backfill-tool run \
  --collection collection.json \
  --csv data.csv \
  --threads 10 \
  --batch-size 1000
```

## 📖 Usage

### Basic Command

```bash
backfill-tool run [flags]
```

### Flags

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--collection` | `-c` | Path to Postman collection JSON file | - | Yes |
| `--csv` | `-s` | Path to CSV data file | - | Yes |
| `--threads` | `-t` | Number of concurrent worker threads | 10 | No |
| `--batch-size` | `-b` | Number of records per batch | 1000 | No |
| `--verbose` | `-v` | Enable verbose output | false | No |

### Example Commands

```bash
# Basic usage with 10 threads
./backfill-tool run -c collection.json -s data.csv -t 10

# High concurrency for faster processing
./backfill-tool run -c collection.json -s data.csv -t 50

# Low concurrency for rate-limited APIs
./backfill-tool run -c collection.json -s data.csv -t 2

# With verbose output
./backfill-tool run -c collection.json -s data.csv -t 10 -v
```

## 💡 Examples

### Example 1: Simple POST Request with Body Variables

**CSV File** (`users.csv`):
```csv
name,email,age
John Doe,john@example.com,30
Jane Smith,jane@example.com,25
```

**Postman Collection Request**:
```json
{
  "name": "Create User",
  "request": {
    "method": "POST",
    "url": {
      "raw": "https://api.example.com/users"
    },
    "header": [
      {
        "key": "Content-Type",
        "value": "application/json"
      }
    ],
    "body": {
      "mode": "raw",
      "raw": "{\n  \"name\": \"{{name}}\",\n  \"email\": \"{{email}}\",\n  \"age\": {{age}}\n}"
    }
  }
}
```

**Result**: Creates two users with the data from CSV

---

### Example 2: GET Request with Path Variables

**CSV File** (`user-ids.csv`):
```csv
userId
1
2
3
```

**Postman Collection Request**:
```json
{
  "name": "Get User by ID",
  "request": {
    "method": "GET",
    "url": {
      "raw": "https://api.example.com/users/{{userId}}"
    }
  }
}
```

**Result**: Fetches users with IDs 1, 2, and 3

---

### Example 3: Query Parameters

**CSV File** (`search.csv`):
```csv
query,limit,offset
smartphones,10,0
laptops,20,0
tablets,15,10
```

**Postman Collection Request**:
```json
{
  "name": "Search Products",
  "request": {
    "method": "GET",
    "url": {
      "raw": "https://api.example.com/products?q={{query}}&limit={{limit}}&offset={{offset}}"
    }
  }
}
```

**Result**: Executes three search queries with different parameters

---

### Example 4: Dynamic Headers

**CSV File** (`auth-requests.csv`):
```csv
token,userId
eyJhbGc...,user123
eyJhbGc...,user456
```

**Postman Collection Request**:
```json
{
  "name": "Authenticated Request",
  "request": {
    "method": "GET",
    "url": {
      "raw": "https://api.example.com/users/{{userId}}/profile"
    },
    "header": [
      {
        "key": "Authorization",
        "value": "Bearer {{token}}"
      }
    ]
  }
}
```

**Result**: Makes authenticated requests with different tokens

---

### Example 5: Nested Folders

**Postman Collection Structure**:
```
📁 User Management
  📁 CRUD Operations
    🔧 Create User (POST)
    🔧 Get User (GET)
    🔧 Update User (PUT)
    🔧 Delete User (DELETE)
  📁 Admin Operations
    🔧 Get All Users (GET)
    🔧 Bulk Delete (DELETE)
```

**Result**: The tool processes all requests in all nested folders recursively

---

### Example 6: Multiple HTTP Methods

**CSV File** (`operations.csv`):
```csv
userId,name,email
1,John Updated,john.new@example.com
2,Jane Updated,jane.new@example.com
```

**Postman Collection** with multiple methods:
- POST to create
- GET to verify
- PUT to update
- DELETE to cleanup

**Result**: Executes all operations for each CSV row

## 🔤 Template Variable Syntax

The tool uses `{{variableName}}` syntax for template variables. Variables are replaced with values from the corresponding CSV column.

### Supported Locations

1. **URL Path**:
   ```
   https://api.example.com/users/{{userId}}/posts/{{postId}}
   ```

2. **Query Parameters**:
   ```
   https://api.example.com/search?q={{query}}&type={{type}}&limit={{limit}}
   ```

3. **Headers**:
   ```json
   {
     "key": "Authorization",
     "value": "Bearer {{token}}"
   }
   ```

4. **Request Body** (JSON):
   ```json
   {
     "name": "{{name}}",
     "email": "{{email}}",
     "metadata": {
       "source": "{{source}}"
     }
   }
   ```

5. **Request Body** (Text/Plain):
   ```
   User {{name}} with email {{email}} registered in {{year}}
   ```

### Variable Matching

- Variable names are **case-sensitive** and must match CSV column headers exactly
- If a variable is not found in CSV data, it remains unchanged as `{{variableName}}`
- Whitespace inside brackets is trimmed: `{{ name }}` = `{{name}}`

## ⚙️ Configuration Options

### Threads (`--threads` / `-t`)

Controls the number of concurrent worker goroutines. Higher values increase throughput but may overwhelm the target API.

**Guidelines**:
- **Low (1-5)**: For rate-limited APIs or when order matters
- **Medium (10-20)**: Balanced performance for most APIs
- **High (50-100)**: For high-throughput APIs and bulk operations
- **Very High (100+)**: For internal APIs or load testing

**Example**:
```bash
# Conservative approach for external API
./backfill-tool run -c collection.json -s data.csv -t 5

# Aggressive approach for internal API
./backfill-tool run -c collection.json -s data.csv -t 100
```

### Batch Size (`--batch-size` / `-b`)

Currently informational. Reserved for future batch processing features.

### Verbose Mode (`--verbose` / `-v`)

Enables detailed logging including:
- Request/response details
- Worker activity
- Timing information

## 📊 CSV File Format

### Requirements

1. **First row must be headers**: Column names used for template variables
2. **Consistent columns**: All rows should have the same number of columns
3. **Encoding**: UTF-8 encoding recommended
4. **Delimiter**: Comma (`,`) as the default delimiter

### Example CSV

```csv
id,name,email,age,country
1,Alice Smith,alice@example.com,30,USA
2,Bob Johnson,bob@example.com,25,Canada
3,Charlie Brown,charlie@example.com,35,UK
```

### Tips

- Use quotes for fields containing commas: `"Smith, John"`
- Empty fields are supported: `1,John,,30` (email is empty)
- Avoid special characters in header names (use `user_id` instead of `user-id`)

## 📮 Postman Collection Setup

### Exporting from Postman

1. Open Postman
2. Select your collection
3. Click the three dots (⋯) next to the collection name
4. Select **Export**
5. Choose **Collection v2.1** format
6. Save the JSON file

### Collection Best Practices

1. **Organize with folders**: Group related requests
2. **Use descriptive names**: Name requests clearly
3. **Add documentation**: Use Postman's description fields
4. **Test before export**: Ensure requests work in Postman
5. **Use variables consistently**: Match CSV column names

### Supported Collection Features

✅ **Supported**:
- All HTTP methods (GET, POST, PUT, PATCH, DELETE, etc.)
- Headers with variables
- URL path variables
- Query parameters with variables
- JSON request bodies
- Plain text request bodies
- Nested folders (unlimited depth)

❌ **Not Yet Supported**:
- Postman environment variables
- Pre-request scripts
- Tests/assertions
- Form data (multipart/form-data)
- File uploads
- Authentication helpers (OAuth, etc.)

## 🎯 Advanced Usage

### Scenario 1: API Migration

Migrate data from one system to another:

```bash
# Export data from old system to CSV
# Prepare Postman collection for new system API
./backfill-tool run -c new-api-collection.json -s exported-data.csv -t 20
```

### Scenario 2: Load Testing

Generate load on your API:

```bash
# Create CSV with test data
# Use high thread count
./backfill-tool run -c load-test.json -s test-data.csv -t 100
```

### Scenario 3: Data Validation

Verify data across systems:

```bash
# Low thread count for careful verification
./backfill-tool run -c validation.json -s data-to-verify.csv -t 2 -v
```

### Scenario 4: Batch Updates

Update multiple records:

```bash
# CSV with IDs and new values
./backfill-tool run -c update-collection.json -s updates.csv -t 15
```

## 🔍 Troubleshooting

### Issue: "Error opening collection file"

**Solution**: Check that the file path is correct and the file exists
```bash
# Verify file exists
ls -la collection.json

# Use absolute path if needed
./backfill-tool run -c /full/path/to/collection.json -s data.csv -t 10
```

### Issue: "Error parsing collection JSON"

**Solution**: Validate your Postman collection JSON
```bash
# Check JSON validity
cat collection.json | jq .

# Re-export from Postman if invalid
```

### Issue: "CSV file is empty"

**Solution**: Ensure CSV has headers and at least one data row
```bash
# Check CSV content
head data.csv

# Verify it has at least 2 lines (header + data)
wc -l data.csv
```

### Issue: Variables not being replaced

**Solution**:
1. Verify CSV column names match variable names exactly (case-sensitive)
2. Check for typos in `{{variableName}}`
3. Ensure CSV file is properly formatted

### Issue: Too many failed requests

**Solution**:
1. Reduce thread count to avoid overwhelming the API
2. Check API rate limits
3. Verify request format in Postman first
4. Use verbose mode to see error details: `-v`

### Issue: Requests timing out

**Solution**:
- The default timeout is 30 seconds
- Check network connectivity
- Verify API endpoint is responsive
- Reduce concurrent threads

## 📈 Performance Tips

1. **Optimize Thread Count**:
   - Start with 10 threads
   - Increase gradually while monitoring API response times
   - Watch for rate limiting errors

2. **CSV File Size**:
   - The tool loads the entire CSV into memory
   - For very large files (>1M rows), consider splitting into smaller batches

3. **Request Complexity**:
   - Simple GET requests: Higher thread counts (50-100)
   - Complex POST/PUT with large payloads: Lower thread counts (10-20)

4. **Network Considerations**:
   - Local/internal APIs: Higher thread counts
   - External APIs: Respect rate limits, use lower thread counts

## 🏗️ Project Structure

```
backfill-tool/
├── main.go                   # Application entry point
├── cmd/
│   ├── root.go              # Root CLI command
│   └── run.go               # Run command with flags
├── internal/
│   └── run_batch.go         # Core batch processing logic
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── example.csv              # Simple example CSV
├── example-data.csv         # Comprehensive example CSV
├── example-collection.json  # Example Postman collection
├── README.md                # This file
└── LICENSE                  # License file
```

## 🧪 Testing

Run the included examples:

```bash
# Test with example files
./backfill-tool run \
  -c example-collection.json \
  -s example-data.csv \
  -t 5
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/sukhmanjit-singh/backfill-tool.git
cd backfill-tool

# Install dependencies
go mod download

# Build
go build -o backfill-tool .

# Run tests (when available)
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra CLI](https://github.com/spf13/cobra)
- Inspired by the need for efficient bulk API operations
- Thanks to the Postman team for the excellent collection format

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/sukhmanjit-singh/backfill-tool/issues)
- **Discussions**: [GitHub Discussions](https://github.com/sukhmanjit-singh/backfill-tool/discussions)

## 🗺️ Roadmap

- [ ] Support for Postman environment variables
- [ ] Response validation and assertions
- [ ] Export results to CSV/JSON
- [ ] Retry logic for failed requests
- [ ] Progress bar for large batches
- [ ] Support for authentication flows
- [ ] Docker image for easy deployment
- [ ] Rate limiting configuration
- [ ] Request scheduling and delays

---

**Made with ❤️ for the API automation community**

## 📊 Metrics & Analytics

### Automatic Metrics Export

Every run automatically generates a metrics JSON file with comprehensive statistics:

```json
{
  "collection_name": "User Management API",
  "csv_file": "data.csv",
  "start_time": "2025-11-03T14:30:00Z",
  "end_time": "2025-11-03T14:32:30Z",
  "duration_seconds": 150.5,
  "total_records": 1000,
  "summary": {
    "total_requests": 3000,
    "successful": 2950,
    "failed": 50,
    "success_rate_pct": 98.33
  },
  "items": [
    {
      "name": "Create User",
      "total_requests": 1000,
      "successful": 987,
      "failed": 13,
      "success_rate_pct": 98.70,
      "timing": {
        "avg_ms": 145,
        "min_ms": 89,
        "max_ms": 2300
      },
      "duration_seconds": 145.2
    }
  ]
}
```

### Custom Metrics Location

Specify a custom path for metrics:

```bash
# Save to specific location
./backfill-tool run -c collection.json -s data.csv --metrics-file ./results/run-$(date +%Y%m%d).json

# Default location: metrics_YYYYMMDD_HHMMSS.json
```

## 🔄 Failed Request Workflow

### How It Works

1. **Automatic Detection**: Failed requests (non-2xx status codes, timeouts, errors) are tracked
2. **CSV Export**: Saved in same format as input CSV for easy retry
3. **One-Command Retry**: Use the failed CSV directly

### Example Workflow

```bash
# Initial run
./backfill-tool run -c collection.json -s data.csv -t 10

# Output shows:
# ❌ Failed requests saved to: failed_requests_Create_User_20251103_143000.csv

# Retry just the failed requests
./backfill-tool run -c collection.json -s failed_requests_Create_User_20251103_143000.csv -t 5

# Repeat until all succeed or investigate errors
```

### Failed Request CSV Format

The CSV contains exactly the same columns as your input CSV:

**Original CSV** (`users.csv`):
```csv
userId,name,email
1,John,john@example.com
2,Jane,jane@example.com
3,Bob,bob@example.com
```

**Failed Requests CSV** (`failed_requests_Create_User_20251103_143000.csv`):
```csv
userId,name,email
2,Jane,jane@example.com
3,Bob,bob@example.com
```

**Retry**:
```bash
./backfill-tool run -c collection.json -s failed_requests_Create_User_20251103_143000.csv -t 5
```

## 📈 Real-time Progress Display

### Progress Bar Features

The live progress bar shows:
- **Completion**: Visual bar and percentage
- **Success/Failure**: Live counts with colored indicators (✓ green, ✗ red)
- **Performance**: Average response time in milliseconds
- **ETA**: Estimated time to completion

### Example Output

```
Progress: [████████████░░░░░░░░] 45/100 (45.0%) | ✓42 ✗3 | Avg: 234ms | ETA: 12s
```

### Quiet Mode for CI/CD

Suppress progress bars while still saving metrics:

```bash
# Perfect for Jenkins, GitHub Actions, etc.
./backfill-tool run -c collection.json -s data.csv -t 20 --quiet
```

## 🎨 Output Examples

### Standard Output (Normal Mode)

```
🚀 Backfill Tool v2.1.0
📦 Collection: example-collection.json
📊 CSV Data: example-data.csv
⚙️  Workers: 10

📦 Collection: User Management API
📊 Items found: 3
📂 Reading CSV file: data.csv
✓ Loaded 1000 records from CSV

🔧 Processing: Create User
   Method: POST | URL: https://api.example.com/users
   Records: 1000 | Workers: 10

Progress: [████████████████████] 1000/1000 (100%) | ✓987 ✗13 | Avg: 145ms | ETA: 0s
   ❌ Failed requests saved to: failed_requests_Create_User_20251103_143000.csv

📊 Summary:
   Total:        1000
   Successful:   987 (98.7%)
   Failed:       13 (1.3%)
   Avg Time:     145ms
   Min Time:     89ms
   Max Time:     2.3s
   Duration:     2m 25s

💾 Metrics saved to: metrics_20251103_143000.json

============================================================
🎯 EXECUTION COMPLETE
============================================================
Collection:     User Management API
Total Requests: 3000
Successful:     2950 (98.3%)
Failed:         50 (1.7%)
Duration:       7m 30s
Throughput:     6.67 req/s
============================================================
```

### Quiet Mode Output (CI/CD)

```bash
./backfill-tool run -c collection.json -s data.csv --quiet
# (No output except errors)
# Metrics still saved to metrics_YYYYMMDD_HHMMSS.json
```

## 🤖 CI/CD Integration

### GitHub Actions Example

```yaml
name: Data Backfill

on:
  workflow_dispatch:
    inputs:
      csv_file:
        description: 'CSV file to process'
        required: true

jobs:
  backfill:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install Backfill Tool
        run: |
          git clone https://github.com/sukhmanjit-singh/backfill-tool.git
          cd backfill-tool
          go build -o backfill-tool .
      
      - name: Run Backfill
        run: |
          ./backfill-tool/backfill-tool run \
            --collection ./collections/api.json \
            --csv ${{ github.event.inputs.csv_file }} \
            --threads 20 \
            --quiet \
            --metrics-file ./metrics.json
      
      - name: Upload Metrics
        uses: actions/upload-artifact@v3
        with:
          name: metrics
          path: metrics.json
      
      - name: Upload Failed Requests
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: failed-requests
          path: failed_requests_*.csv
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any
    
    parameters {
        string(name: 'CSV_FILE', description: 'CSV file path')
        string(name: 'THREADS', defaultValue: '10', description: 'Worker threads')
    }
    
    stages {
        stage('Backfill') {
            steps {
                script {
                    sh """
                        ./backfill-tool run \
                            -c ./collections/api.json \
                            -s ${params.CSV_FILE} \
                            -t ${params.THREADS} \
                            --quiet \
                            --metrics-file ./metrics.json
                    """
                }
            }
        }
        
        stage('Archive Results') {
            steps {
                archiveArtifacts artifacts: 'metrics.json', allowEmptyArchive: false
                archiveArtifacts artifacts: 'failed_requests_*.csv', allowEmptyArchive: true
            }
        }
    }
}
```

## 💡 Pro Tips

### 1. Optimal Thread Count

```bash
# Start conservative
./backfill-tool run -c collection.json -s data.csv -t 5

# Monitor avg response time in progress bar
# If stable and fast, increase threads

./backfill-tool run -c collection.json -s data.csv -t 20

# Watch for rate limiting (sudden spikes in response time)
```

### 2. Retry Strategy

```bash
# First attempt with normal thread count
./backfill-tool run -c collection.json -s data.csv -t 10

# Retry failures with lower threads (might be rate limiting)
./backfill-tool run -c collection.json -s failed_requests_*.csv -t 2

# Final retry with minimal threads
./backfill-tool run -c collection.json -s failed_requests_*.csv -t 1
```

### 3. Metrics Analysis

```bash
# Extract success rate from metrics
cat metrics_*.json | jq '.summary.success_rate_pct'

# Find slowest requests
cat metrics_*.json | jq '.items | sort_by(.timing.max_ms) | reverse | .[0]'

# Compare two runs
diff <(cat metrics_run1.json | jq '.summary') <(cat metrics_run2.json | jq '.summary')
```

### 4. Debugging Failed Requests

```bash
# Run with verbose mode for detailed logging
./backfill-tool run -c collection.json -s failed_requests_*.csv -t 1 --verbose

# Check specific failure patterns
cat metrics_*.json | jq '.items[] | select(.failed > 0)'
```

## 📝 Command Reference

### Main Commands

```bash
# Run backfill with defaults
backfill-tool run -c collection.json -s data.csv

# Show all examples
backfill-tool examples

# Show version and features
backfill-tool version

# Get help for any command
backfill-tool run --help
```

### Common Flag Combinations

```bash
# High performance mode
backfill-tool run -c collection.json -s data.csv -t 50

# Careful/debug mode
backfill-tool run -c collection.json -s data.csv -t 1 --verbose

# CI/CD mode
backfill-tool run -c collection.json -s data.csv -t 20 --quiet --metrics-file ./output/metrics.json

# Retry failures conservatively
backfill-tool run -c collection.json -s failed_requests_*.csv -t 2
```


## 🔗 Query Parameters - Complete Guide

### Overview

The backfill tool supports comprehensive query parameter replacement from CSV data. Query parameters can be specified in two ways:

1. **Raw URL template** - Parameters directly in the URL string
2. **Structured parameters** - Postman's query parameter array

Both methods support template variable replacement with proper URL encoding.

### Method 1: Raw URL Template

Simply include `{{variableName}}` placeholders in your URL query string:

**Postman Collection:**
```json
{
  "url": {
    "raw": "https://api.example.com/search?q={{query}}&limit={{limit}}&offset={{offset}}"
  }
}
```

**CSV File:**
```csv
query,limit,offset
smartphones,10,0
tablets,20,10
laptops,15,20
```

**Generated URLs:**
- `https://api.example.com/search?q=smartphones&limit=10&offset=0`
- `https://api.example.com/search?q=tablets&limit=20&offset=10`
- `https://api.example.com/search?q=laptops&limit=15&offset=20`

### Method 2: Structured Query Parameters (Postman)

Use Postman's query parameter array for better organization:

**Postman Collection:**
```json
{
  "url": {
    "raw": "https://api.example.com/users",
    "query": [
      {
        "key": "name",
        "value": "{{name}}"
      },
      {
        "key": "email",
        "value": "{{email}}"
      },
      {
        "key": "status",
        "value": "active"
      }
    ]
  }
}
```

**CSV File:**
```csv
name,email
Alice Smith,alice@example.com
Bob Johnson,bob@example.com
```

**Generated URLs:**
- `https://api.example.com/users?email=alice%40example.com&name=Alice+Smith&status=active`
- `https://api.example.com/users?email=bob%40example.com&name=Bob+Johnson&status=active`

Note: Email addresses are properly URL-encoded (`@` becomes `%40`).

### Combining Path Variables and Query Parameters

You can use both path variables and query parameters together:

**Postman Collection:**
```json
{
  "url": {
    "raw": "https://api.example.com/users/{{userId}}/posts?limit={{limit}}&tag={{tag}}"
  }
}
```

**CSV File:**
```csv
userId,limit,tag
123,10,urgent
456,20,normal
```

**Generated URLs:**
- `https://api.example.com/users/123/posts?limit=10&tag=urgent`
- `https://api.example.com/users/456/posts?limit=20&tag=normal`

### Special Characters & URL Encoding

The tool automatically handles URL encoding for special characters:

**CSV File:**
```csv
searchTerm,filter
camera & lens,type:dslr
laptop (15 inch),brand:dell
```

**Generated Query String:**
```
?searchTerm=camera+%26+lens&filter=type%3Adslr
?searchTerm=laptop+%2815+inch%29&filter=brand%3Adell
```

Special characters are encoded:
- Space → `+` or `%20`
- `&` → `%26`
- `=` → `%3D`
- `(` → `%28`
- `)` → `%29`
- `:` → `%3A`
- `@` → `%40`

### Multiple Query Parameters with Same Name

Some APIs accept multiple values for the same parameter:

**Postman Collection:**
```json
{
  "url": {
    "raw": "https://api.example.com/filter?category={{category1}}&category={{category2}}"
  }
}
```

**CSV File:**
```csv
category1,category2
tech,gadgets
hardware,software
```

**Generated URLs:**
- `https://api.example.com/filter?category=tech&category=gadgets`
- `https://api.example.com/filter?category=hardware&category=software`

### Query Parameters with POST/PUT Requests

Query parameters work with all HTTP methods:

**Postman Collection:**
```json
{
  "method": "POST",
  "url": {
    "raw": "https://api.example.com/users?notify={{notify}}&async={{async}}"
  },
  "body": {
    "raw": "{\"name\": \"{{name}}\"}"
  }
}
```

**CSV File:**
```csv
name,notify,async
Alice,true,false
Bob,false,true
```

**Result:**
- POST to `https://api.example.com/users?notify=true&async=false` with body `{"name": "Alice"}`
- POST to `https://api.example.com/users?notify=false&async=true` with body `{"name": "Bob"}`

### Testing Query Parameters

Use the included test files to verify query parameter functionality:

```bash
# Test basic query parameters
./backfill-tool run -c test-query-params.json -s test-query-params.csv -t 2

# The test collection includes:
# - Simple query parameters in raw URL
# - Structured query parameters (Postman array)
# - Mixed path and query parameters
# - Special character encoding
# - POST requests with query params
```

### Best Practices

1. **Use Structured Parameters** for better readability in Postman
2. **URL Encode in CSV** is NOT needed - the tool handles encoding automatically
3. **Empty Values** are supported - empty CSV fields result in empty query param values
4. **Static Parameters** can be mixed with dynamic ones (e.g., `status=active` + `name={{name}}`)
5. **Test in Postman First** - verify your collection works before bulk execution

### Common Patterns

#### Pagination
```csv
page,pageSize
1,100
2,100
3,100
```
URL: `/api/data?page={{page}}&pageSize={{pageSize}}`

#### Filtering
```csv
startDate,endDate,status
2025-01-01,2025-01-31,active
2025-02-01,2025-02-28,pending
```
URL: `/api/records?startDate={{startDate}}&endDate={{endDate}}&status={{status}}`

#### Search with Filters
```csv
query,category,minPrice,maxPrice
laptops,electronics,500,2000
phones,mobile,200,1500
```
URL: `/api/search?q={{query}}&category={{category}}&minPrice={{minPrice}}&maxPrice={{maxPrice}}`

### Troubleshooting

**Query params not being replaced?**
- Check CSV column names match exactly (case-sensitive)
- Verify `{{variableName}}` syntax is correct
- Try verbose mode to see generated URLs

**Special characters not encoded?**
- This is handled automatically - no action needed
- If you pre-encode in CSV, it will be double-encoded

**Parameters in wrong order?**
- Query parameter order may vary (this is normal and doesn't affect functionality)
- APIs should accept parameters in any order

## 📋 Enhanced Failed Request Logging

### Overview

When requests fail, the tool automatically saves them to a CSV file with **complete error details**. The CSV is designed to serve two purposes:

1. **Easy Retry** - Re-upload the same CSV to retry failed requests
2. **Error Analysis** - View detailed error information for debugging

### CSV Format

The failed requests CSV contains:

**Original Data Columns** (from your input CSV)
- All original columns preserved exactly
- Same order as input CSV
- Can be directly re-uploaded

**Error Detail Columns** (added automatically)
- `_error_status_code` - HTTP status code (200, 404, 500, etc.) or 0 for connection errors
- `_error_message` - Detailed error message from the API
- `_error_url` - Complete URL that was called (with variables replaced)
- `_error_method` - HTTP method used (GET, POST, PUT, etc.)
- `_error_timestamp` - When the request failed (RFC3339 format)
- `_error_response_time_ms` - How long the request took in milliseconds

### Example Failed Requests CSV

```csv
name,year,userId,email,_error_status_code,_error_message,_error_url,_error_method,_error_timestamp,_error_response_time_ms
Alice Smith,2020,1,alice@example.com,500,HTTP 500: Internal Server Error,https://api.example.com/users,POST,2025-11-04T10:30:45Z,1234
Bob Johnson,2021,2,bob@example.com,404,HTTP 404: User not found,https://api.example.com/users,POST,2025-11-04T10:30:46Z,234
Charlie Brown,2019,3,charlie@example.com,0,Request failed: connection refused,https://api.example.com/users,POST,2025-11-04T10:30:47Z,5000
```

### How To Use

#### 1. View Error Details

Open the CSV in Excel, Google Sheets, or any spreadsheet application:

- **Sort by status code** to group similar errors
- **Filter by error message** to find patterns
- **Check response times** to identify slow requests
- **Review URLs** to verify correct variable replacement

#### 2. Retry Failed Requests

Simply re-upload the same CSV - error columns are automatically ignored:

```bash
# First run - some failures
./backfill-tool run -c collection.json -s data.csv -t 10

# Output shows:
# ❌ Failed: 13 requests saved to failed_requests_Create_User_20251104_103045.csv

# Retry just the failed ones
./backfill-tool run -c collection.json -s failed_requests_Create_User_20251104_103045.csv -t 5
```

The tool ignores the `_error_*` columns because they don't match any template variables in your collection.

#### 3. Analyze Patterns

```bash
# Count errors by status code
cut -d',' -f6 failed_requests_*.csv | sort | uniq -c

# Find all 500 errors
grep ",500," failed_requests_*.csv

# Check which URLs failed most
cut -d',' -f8 failed_requests_*.csv | sort | uniq -c | sort -rn
```

### Error Status Codes

| Status Code | Meaning | Common Causes |
|-------------|---------|---------------|
| 0 | Connection Error | Network issues, DNS failure, timeout |
| 400 | Bad Request | Invalid data format, missing required fields |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Valid auth but insufficient permissions |
| 404 | Not Found | Resource doesn't exist (check path variables) |
| 429 | Too Many Requests | Rate limiting - reduce thread count |
| 500 | Internal Server Error | API server issue - retry later |
| 502 | Bad Gateway | Proxy/load balancer issue |
| 503 | Service Unavailable | API temporarily down |
| 504 | Gateway Timeout | Request took too long |

### Common Error Messages

**Connection Errors (Status Code = 0)**
- `Request failed: connection refused` - Service not running
- `Request failed: timeout` - Request took longer than 30 seconds
- `Request failed: no such host` - DNS resolution failed
- `Request failed: EOF` - Connection closed unexpectedly

**HTTP Errors (Status Code > 0)**
- `HTTP 400: Bad Request` - Check request body format
- `HTTP 401: Unauthorized` - Add authentication headers
- `HTTP 404: Not found` - Verify URL path variables
- `HTTP 500: Internal Server Error` - Server-side issue

### Retry Strategies

#### Strategy 1: Retry All Failures

```bash
# Simple retry with lower concurrency
./backfill-tool run -c collection.json -s failed_requests_*.csv -t 2
```

#### Strategy 2: Filter by Error Type

```bash
# Extract only timeout errors (0 status or took >5000ms)
awk -F',' '$6 == 0 || $11 > 5000' failed_requests_Create_User_20251104.csv > timeouts.csv

# Retry just timeouts with minimal concurrency
./backfill-tool run -c collection.json -s timeouts.csv -t 1
```

#### Strategy 3: Retry 5xx Errors Only

```bash
# Server errors might succeed on retry
awk -F',' '$6 >= 500 && $6 < 600' failed_requests_*.csv > server_errors.csv

# Retry server errors
./backfill-tool run -c collection.json -s server_errors.csv -t 5
```

#### Strategy 4: Fix Data and Retry 4xx Errors

```bash
# 4xx errors usually indicate data problems
# Review error messages first
grep ",4" failed_requests_*.csv | less

# Fix the data in CSV
# Then retry
./backfill-tool run -c collection.json -s fixed_data.csv -t 10
```

### Best Practices

1. **Always Review Errors** - Check the CSV before retrying
2. **Reduce Threads for Retries** - Failed requests often indicate issues; go slower
3. **Look for Patterns** - If all failures have the same error, fix the root cause
4. **Check Status Codes** - 4xx = client issue, 5xx = server issue
5. **Monitor Response Times** - High response times might indicate API strain

### Troubleshooting

**Error columns are empty?**
- Request failed before getting a response
- Check _error_message for connection errors

**Can't re-upload CSV?**
- Make sure you're uploading to the same collection
- Error columns are harmless - they'll be ignored

**Too many retries?**
- If retrying doesn't help after 2-3 attempts, investigate the root cause
- Check metrics.json for patterns
