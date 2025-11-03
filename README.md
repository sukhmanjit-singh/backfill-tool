# Backfill Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A high-performance CLI tool for executing bulk API requests using Postman collections and CSV data. Perfect for data migration, backfilling, bulk testing, and API automation tasks.

## 🚀 Features

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
