# Testing Guide

The literature project uses `gotestsum` for enhanced test output and follows comprehensive testing practices. This guide covers all aspects of testing within the project.

## Quick Start

```bash
# Run all tests with clean output
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis

# Run specific test pattern
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -run TestFindSimilar ./...

# Run with verbose output for debugging
gotestsum --format-hide-empty-pkg --format standard-verbose --format-icons hivis
```

## Test Commands

### Basic Commands

```bash
# Run all tests with clean output
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis

# Run specific test pattern
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -run TestFetchArticle ./...

# Run with verbose output for debugging
gotestsum --format-hide-empty-pkg --format standard-verbose --format-icons hivis

# Run tests for specific package
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis ./internal/...

# Run tests with coverage
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -cover ./...

# Run tests with race detection
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -race ./...
```

### Advanced Commands

```bash
# Generate coverage report
gotestsum -- -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total:

# Run benchmarks for performance testing
go test -bench=. ./...

# Test with race condition detection
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -race ./...
```

## Testing Guidelines

### Test Structure

Follow the Arrange-Act-Assert pattern for test organization:

```go
func TestSearchPubMed(t *testing.T) {
    // Arrange
    client, err := literature.New()
    require.NoError(t, err)
    query := "machine learning"

    // Act
    result, err := client.Search(query)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.PMIDs)
}
```

### Test Types

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test interaction between components
3. **External API Tests**: Test actual API calls (use build tags for optional execution)

```go
//go:build integration
// +build integration

func TestSearchPubMedIntegration(t *testing.T) {
    // Integration test that makes real API calls
    client, err := literature.New()
    require.NoError(t, err)
    
    result, err := client.Search("COVID-19")
    assert.NoError(t, err)
    assert.NotEmpty(t, result.PMIDs)
}
```

### Mock Usage

Use testify/mock for mocking external dependencies:

```go
type MockHTTPClient struct {
    mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    args := m.Called(req)
    return args.Get(0).(*http.Response), args.Error(1)
}

func TestWithMockClient(t *testing.T) {
    mockClient := new(MockHTTPClient)
    client, err := literature.New(
        literature.WithHTTPClient(mockClient),
    )
    require.NoError(t, err)
    
    // Set up expectations
    mockResponse := &http.Response{
        StatusCode: 200,
        Body: io.NopCloser(strings.NewReader(`<xml>mock response</xml>`)),
    }
    mockClient.On("Do", mock.Anything).Return(mockResponse, nil)
    
    // Test execution
    result, err := client.Search("test query")
    
    // Assertions
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestValidatePMID(t *testing.T) {
    tests := []struct {
        name    string
        pmid    string
        wantErr bool
        errType ErrorType
    }{
        {
            name:    "valid PMID",
            pmid:    "12345678",
            wantErr: false,
        },
        {
            name:    "empty PMID",
            pmid:    "",
            wantErr: true,
            errType: ErrorTypeInvalidInput,
        },
        {
            name:    "non-numeric PMID",
            pmid:    "abc123",
            wantErr: true,
            errType: ErrorTypeInvalidInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePMID(tt.pmid)
            
            if tt.wantErr {
                assert.Error(t, err)
                if litErr, ok := err.(*Error); ok {
                    assert.Equal(t, tt.errType, litErr.Type)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Test Fixtures

Use the `testdata/` directory for test fixtures:

```go
func loadTestData(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    return data
}

func TestParseArticleXML(t *testing.T) {
    xmlData := loadTestData(t, "sample_article.xml")
    
    article, err := ParseArticleXML(xmlData)
    
    assert.NoError(t, err)
    assert.Equal(t, "Sample Title", article.Title)
}
```

### Benchmarking

Include benchmarks for performance-critical functions:

```go
func BenchmarkMapWithError(b *testing.B) {
    pmids := make([]string, 100)
    for i := range pmids {
        pmids[i] = fmt.Sprintf("%d", i+1000000)
    }
    
    processFunc := func(pmid string) (*Article, error) {
        // Simulate processing
        time.Sleep(time.Microsecond)
        return &Article{}, nil
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = MapWithError(pmids, processFunc)
    }
}
```

### Error Testing

Test error conditions explicitly:

```go
func TestSearchWithNetworkError(t *testing.T) {
    // Arrange
    mockClient := &MockHTTPClient{}
    client, err := literature.New(
        literature.WithHTTPClient(mockClient),
    )
    require.NoError(t, err)
    
    networkErr := errors.New("network error")
    mockClient.On("Do", mock.Anything).Return((*http.Response)(nil), networkErr)
    
    // Act
    result, err := client.Search("test query")
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "network error")
}
```

### Test Coverage

Aim for high test coverage while focusing on meaningful tests:

```bash
# Generate coverage report
gotestsum -- -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total:
```

#### Coverage Goals

- **Unit Tests**: Aim for 90%+ coverage on core business logic
- **Integration Tests**: Cover happy path and common error scenarios
- **Edge Cases**: Test boundary conditions and error paths

## Test Organization

```
literature/
├── internal/
│   ├── search/
│   │   ├── search.go
│   │   ├── search_test.go
│   │   └── search_integration_test.go
│   └── article/
│       ├── article.go
│       └── article_test.go
├── testdata/
│   ├── sample_article.xml
│   ├── search_result.xml
│   └── error_response.xml
└── examples/
    └── example_test.go
```

### File Naming Conventions

- `*_test.go` - Unit tests
- `*_integration_test.go` - Integration tests
- `*_bench_test.go` - Benchmark tests (optional, can be in regular test files)
- `example_*_test.go` - Example tests for documentation

### Test Data Organization

```
testdata/
├── articles/
│   ├── valid_article.xml
│   ├── article_with_doi.xml
│   └── article_without_abstract.xml
├── search/
│   ├── search_results.xml
│   ├── empty_results.xml
│   └── error_response.xml
└── pdf/
    ├── available_pdf.xml
    └── unavailable_pdf.xml
```

## Best Practices

### Test Naming

Use descriptive test names that explain the scenario:

```go
// Good
func TestClient_GetArticle_ValidPMID_ReturnsArticle(t *testing.T) {}
func TestClient_GetArticle_InvalidPMID_ReturnsError(t *testing.T) {}
func TestClient_GetArticle_NetworkError_ReturnsWrappedError(t *testing.T) {}

// Avoid
func TestGetArticle(t *testing.T) {}
func TestGetArticleError(t *testing.T) {}
```

### Test Independence

Each test should be independent and not rely on other tests:

```go
func TestSearchFlow(t *testing.T) {
    // Each subtest is independent
    t.Run("valid_query_returns_results", func(t *testing.T) {
        client := setupTestClient(t)
        // Test implementation
    })
    
    t.Run("empty_query_returns_error", func(t *testing.T) {
        client := setupTestClient(t)
        // Test implementation
    })
}

func setupTestClient(t *testing.T) *literature.Client {
    client, err := literature.New()
    require.NoError(t, err)
    return client
}
```

### Helper Functions

Create helper functions for common test setup:

```go
func createTestArticle(t *testing.T, pmid string) *Article {
    return &Article{
        PMID:     pmid,
        Title:    "Test Article",
        Journal:  "Test Journal",
        PubYear:  "2023",
        Authors:  []string{"Test Author"},
    }
}

func createMockHTTPClient(t *testing.T, response string, statusCode int) *MockHTTPClient {
    mockClient := &MockHTTPClient{}
    resp := &http.Response{
        StatusCode: statusCode,
        Body:       io.NopCloser(strings.NewReader(response)),
    }
    mockClient.On("Do", mock.Anything).Return(resp, nil)
    return mockClient
}
```

### Error Message Testing

Test specific error messages and types:

```go
func TestValidateInput_EmptyPMID_ReturnsSpecificError(t *testing.T) {
    err := ValidatePMID("")
    
    // Test error type
    var validationErr *ValidationError
    assert.True(t, errors.As(err, &validationErr))
    
    // Test error message
    assert.Contains(t, err.Error(), "PMID cannot be empty")
    
    // Test error code if applicable
    assert.Equal(t, "EMPTY_PMID", validationErr.Code)
}
```

## Continuous Integration

### GitHub Actions Integration

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Install gotestsum
      run: go install gotest.tools/gotestsum@latest
    
    - name: Run tests
      run: gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -race ./...
    
    - name: Generate coverage
      run: gotestsum -- -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Pre-commit Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run tests before commit
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

# Check coverage
COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage is below 80%. Current: ${COVERAGE}%"
    exit 1
fi
```

## Troubleshooting

### Common Issues

1. **Tests failing due to network timeouts**:
   ```go
   // Increase timeout for integration tests
   client, err := literature.New(
       literature.WithTimeout(60*time.Second),
   )
   ```

2. **Race conditions in tests**:
   ```bash
   # Run with race detector
   gotestsum -- -race ./...
   ```

3. **Flaky tests**:
   ```bash
   # Run tests multiple times
   gotestsum -- -count=10 ./...
   ```

### Debug Failing Tests

```bash
# Run with verbose output
gotestsum --format-hide-empty-pkg --format standard-verbose --format-icons hivis

# Run specific failing test
gotestsum -- -run TestSpecificFailingTest -v

# Add debug output in tests
func TestSomething(t *testing.T) {
    t.Logf("Debug info: %+v", someVariable)
    // Test implementation
}
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [gotestsum Documentation](https://github.com/gotestyourself/gotestsum)
- [Go Test Coverage](https://go.dev/blog/cover)