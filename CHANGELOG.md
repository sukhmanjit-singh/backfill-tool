# Changelog

All notable changes to the Backfill Tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-11-03

### Added

#### Core Features
- **URL Path Variable Replacement**: Support for `{{variable}}` syntax in URL paths (e.g., `/api/users/{{userId}}`)
- **Query Parameter Replacement**: Template variables now work in query parameters (e.g., `?name={{name}}&id={{id}}`)
- **Header Variable Replacement**: Dynamic header values using CSV data (e.g., `Authorization: Bearer {{token}}`)
- **Nested Folder Support**: Recursive processing of nested folders in Postman collections (unlimited depth)
- **Template Variable System**: Unified `{{variableName}}` replacement across URLs, headers, and bodies

#### Improvements
- **Fixed Thread Parameter**: Worker count now correctly uses the `--threads` flag instead of hardcoded value of 3
- **Optimized CSV Reading**: CSV file is now read once at startup and reused for all requests (major performance improvement)
- **Enhanced Error Handling**: Comprehensive error validation and reporting throughout the application
- **Request Result Tracking**: Detailed success/failure tracking with status codes and response summaries
- **Better Request Identification**: Smart record identification using common fields (id, name, email)

#### Code Quality
- **Production-Ready Code**: Clean, well-documented code with extensive comments
- **Proper Error Messages**: Descriptive error messages with context for troubleshooting
- **Input Validation**: Validates all command-line parameters before execution
- **Type Safety**: Added proper structures for PostmanURL, QueryParam, and RequestResult
- **Better Output Formatting**: Enhanced console output with emojis and indentation for nested structures

#### Documentation
- **Comprehensive README**: Complete documentation with:
  - Detailed feature descriptions
  - Installation instructions
  - Usage examples for all features
  - Template variable syntax guide
  - Troubleshooting section
  - Performance optimization tips
  - Advanced usage scenarios
- **Example Files**:
  - `example-collection.json`: Demonstrates all features (path variables, query params, headers, nested folders)
  - `example-data.csv`: Sample data matching the example collection
- **Code Comments**: Extensive inline documentation for all functions and complex logic

### Changed

#### Breaking Changes
- **JSON Replacement Strategy**: Now uses two-phase replacement:
  1. Direct key matching (backward compatible)
  2. Template variable replacement in string values (new)
- **Error Handling**: Functions now properly propagate errors instead of silently failing

#### Behavior Changes
- **Request Timeout**: Increased from 10 seconds to 30 seconds for better reliability
- **Content-Type Header**: Automatically set to `application/json` when body is present and header not specified
- **Worker Synchronization**: Improved goroutine management with proper cleanup

### Fixed

#### Bugs
- **Format String Bug** (line 94): Fixed missing format specifier in `Printf` statement
- **Redundant Code**: Removed duplicate call to `ReplaceJSONValues`
- **Thread Parameter Ignored**: Fixed hardcoded worker count of 3, now uses `--threads` parameter
- **CSV Re-reading**: Fixed inefficient CSV reading for every collection item
- **Empty Header Handling**: Added validation to skip headers with empty keys or values
- **Missing Error Checks**: Added error handling for JSON decoding in `RunBatch`

#### Edge Cases
- **Empty JSON Bodies**: Proper handling of requests without body content
- **Nested JSON Structures**: Fixed recursive replacement in deeply nested objects and arrays
- **Missing CSV Values**: Gracefully handle missing columns in CSV rows
- **Invalid URLs**: URL validation after template replacement
- **Empty CSV Files**: Better error message when CSV has only headers

### Technical Details

#### Refactored Functions
- `RunBatch()`: Now validates inputs, reads CSV once, and processes recursively
- `processItem()`: New function for recursive folder processing with depth tracking
- `worker()`: Enhanced with comprehensive error handling and result reporting
- `ReplaceJSONValues()`: Improved to handle both direct key matching and template patterns
- `replaceValuesRecursive()`: New name for recursive replacement with better structure handling

#### New Functions
- `processItem()`: Recursively processes Postman items and folders
- `replaceURLVariables()`: Dedicated function for URL template replacement
- `replaceTemplateVariables()`: Generic template variable replacement using regex
- `getRecordInfo()`: Smart extraction of record identifiers for logging

#### New Types
- `QueryParam`: Structure for query parameters
- `RequestResult`: Comprehensive result tracking with success status, status code, and messages

### Dependencies

- Updated minimum Go version requirement to 1.21 (stable release)
- Maintained compatibility with existing dependencies:
  - `github.com/spf13/cobra v1.10.1`
  - `github.com/spf13/pflag v1.0.10`

### Performance Improvements

1. **CSV Loading**: Single read operation instead of per-request reading
2. **Concurrent Processing**: Proper use of the threads parameter for parallelism
3. **Memory Efficiency**: Better goroutine cleanup with WaitGroups
4. **Channel Buffering**: Optimized channel buffer sizes

### Migration Guide from v1.x

#### What stays the same:
- Command-line interface and flags
- CSV file format (headers in first row)
- Postman collection format (v2.1)
- Direct key matching in JSON bodies

#### What's new:
1. **Path variables now work**: Update your collection URLs to use `{{variable}}` syntax
2. **Query parameters supported**: Add variables to query strings
3. **Headers are dynamic**: Use variables in any header value
4. **Nested folders work**: Organize your collection with folders
5. **Template syntax**: Use `{{variableName}}` everywhere, not just in JSON keys

#### Example Migration:

**Before** (v1.x - only JSON body keys):
```json
{
  "body": {
    "raw": "{\"name\": \"fixedValue\", \"userId\": 123}"
  }
}
```

**After** (v2.0 - full template support):
```json
{
  "url": {
    "raw": "https://api.example.com/users/{{userId}}"
  },
  "header": [
    {
      "key": "X-User-Name",
      "value": "{{name}}"
    }
  ],
  "body": {
    "raw": "{\"name\": \"{{name}}\", \"year\": {{year}}}"
  }
}
```

### Security

- No security vulnerabilities introduced
- All HTTP requests use standard library with timeout protection
- No execution of arbitrary code from collections or CSV files
- Input validation on all parameters

### Known Limitations

- Postman environment variables not yet supported
- Pre-request scripts not executed
- Authentication helpers (OAuth) not supported
- Form data and file uploads not implemented
- No retry logic for failed requests

### Acknowledgments

Thanks to the community for feedback on the initial release that led to these improvements.

---

## [1.0.0] - Initial Release

### Features
- Basic Postman collection execution
- CSV data integration
- JSON body variable replacement (direct key matching only)
- Concurrent execution with goroutines
- Command-line interface with Cobra

### Limitations
- Only worked for POST requests with JSON bodies
- No path variable support
- No query parameter support
- No header variable support
- Hardcoded worker count
- No nested folder support
- CSV re-read for every request
