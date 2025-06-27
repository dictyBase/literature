# Literature - Go Client for PubMed

[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](https://github.com/dictyBase/literature/blob/develop/LICENSE)
![GitHub action](https://github.com/dictyBase/literature/workflows/Continuous%20integration/badge.svg)
[![codecov](https://codecov.io/gh/dictyBase/literature/graph/badge.svg?token=RE3OI8BCJS)](https://codecov.io/gh/dictyBase/literature)
![Last commit](https://badgen.net/github/last-commit/dictyBase/literature/develop)
[![Funding](https://badgen.net/badge/Funding/Rex%20L%20Chisholm,dictyBase,DCR/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=10024726&icde=0)

A clean, idiomatic Go library for accessing PubMed literature data through the NCBI eUtils API.

## Features

- 🔍 Search PubMed articles with natural language queries
- 📄 Fetch detailed article metadata by PMID
- 📚 Batch operations for multiple articles
- 🔗 PDF access and availability checking
- ⚙️ Configurable HTTP client with timeout and retry options
- 🛡️ Structured error handling with detailed error types
- 🔄 Thread-safe for concurrent use
- 📖 Comprehensive documentation and examples

## Installation

```bash
go get github.com/dictybase/literature
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/dictybase/literature"
)

func main() {
    // Create a new client
    client, err := literature.New()
    if err != nil {
        log.Fatal(err)
    }

    // Fetch an article by PMID
    article, err := client.GetArticle("12345678")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", article.Title)
    fmt.Printf("Authors: %v\n", article.Authors)
}
```

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Modern API (Recommended)](#modern-api-recommended)
  - [Configuration](#configuration)
  - [Core Methods](#core-methods)
  - [Error Handling](#error-handling)
  - [Examples](#examples)
- [Types and Data Structures](#types-and-data-structures)
- [Error Handling](#error-handling-1)
- [Thread Safety](#thread-safety)
- [Rate Limiting](#rate-limiting)
- [Contributing](#contributing)
- [Development](#development)
  - [Project Structure](#project-structure)
- [Support](#support)

## Modern API (Recommended)

The new `literature` package provides a clean, unified interface for all PubMed operations.

### Configuration

Customize the client with various options:

```go
client, err := literature.New(
    literature.WithTimeout(60*time.Second),
    literature.WithUserAgent("MyApp/1.0"),
    literature.WithHTTPClient(customHTTPClient),
)
```

### Core Methods

- `GetArticle(pmid string) (*Article, error)` - Fetch single article
- `GetArticles(pmids []string) ([]*Article, error)` - Fetch multiple articles
- `Search(query string, opts ...SearchOption) (*SearchResult, error)` - Search articles
- `FindSimilar(pmid string, opts ...SearchOption) (*SearchResult, error)` - Find similar articles
- `GetPDF(pmid string) (*PDF, error)` - Get PDF information
- `HasPDF(pmid string) (bool, error)` - Check PDF availability

#### Configuration Options

- `WithHTTPClient(client *http.Client)` - Custom HTTP client
- `WithTimeout(timeout time.Duration)` - Request timeout
- `WithBaseURL(url string)` - Custom API base URL (for testing)
- `WithUserAgent(userAgent string)` - Custom User-Agent header

#### Search Options

- `WithLimit(limit int)` - Maximum results per search
- `WithOffset(offset int)` - Pagination offset

### Error Handling

The library provides structured error types:

```go
article, err := client.GetArticle("invalid-pmid")
if err != nil {
    if litErr, ok := err.(*literature.Error); ok {
        switch litErr.Type {
        case literature.ErrorTypeInvalidInput:
            // Handle invalid input
        case literature.ErrorTypeArticleNotFound:
            // Handle article not found
        case literature.ErrorTypeNetworkError:
            // Handle network issues
        }
    }
}
```

#### Error Types

- `ErrorTypeInvalidInput` - Invalid input parameters
- `ErrorTypeArticleNotFound` - Article not found
- `ErrorTypePDFNotFound` - PDF not available
- `ErrorTypeNetworkError` - Network-related errors
- `ErrorTypeParseError` - XML parsing errors
- `ErrorTypeAPIError` - PubMed API errors
- `ErrorTypeTimeout` - Request timeouts
- `ErrorTypeRateLimit` - API rate limiting

### Examples

See the `examples/` directory for comprehensive usage examples:

- [`examples/basic/`](examples/basic/) - Basic usage patterns
- [`examples/advanced/`](examples/advanced/) - Advanced configuration and error handling


## Types and Data Structures

### Article Types

#### PubMedArticle

Represents a single PubMed article with comprehensive metadata.

**Key Methods:**
- `GetPMID() string`: Returns PubMed ID
- `GetTitle() string`: Returns article title
- `GetJournalTitle() string`: Returns journal title
- `GetAbstract() string`: Returns article abstract
- `GetAuthors() []Author`: Returns list of authors
- `GetPages() string`: Returns page range
- `GetPubYear() string`: Returns publication year
- `GetPubMonth() string`: Returns publication month
- `GetArticleIDs() []ArticleID`: Returns all article identifiers
- `GetDOI() (string, bool)`: Returns DOI if available
- `GetPMCID() (string, bool)`: Returns PMC ID if available

#### Author

```go
type Author struct {
    LastName string `xml:"LastName"`
    ForeName string `xml:"ForeName"`
}
```

#### ArticleID

```go
type ArticleID struct {
    IDType string `xml:"IdType,attr"`
    Value  string `xml:",chardata"`
}
```

### Search Types

#### ESearchResult

Represents search results from PubMed.

```go
type ESearchResult struct {
    Count    string   // Total number of results
    RetMax   string   // Maximum results returned
    RetStart string   // Starting position
    QueryKey string   // Query key for subsequent requests
    WebEnv   string   // Web environment for subsequent requests
    IDList   struct {
        IDs []string `xml:"Id"`
    }
}
```

**Methods:**
- `GetIDs() []string`: Returns list of PubMed IDs

### PDF Types

#### PDFDownloadInfo

```go
type PDFDownloadInfo struct {
    PMID    string
    PMCID   string
    PDFLink *OALink
}
```

#### OARecord

```go
type OARecord struct {
    ID    string   `xml:"id,attr"`
    Links []OALink `xml:"link"`
}
```

#### OALink

```go
type OALink struct {
    Format string `xml:"format,attr"`
    HREF   string `xml:"href,attr"`
}
```

## Error Handling

### PDFError

Specialized error type for PDF-related operations.

```go
type PDFError struct {
    PMID string
    Type PDFErrorType
    Err  error
}
```

#### PDFErrorType Constants

- `PDFErrorArticleNotFound`: Article metadata not found
- `PDFErrorPMCIDNotFound`: No PMC ID available for article
- `PDFErrorPDFNotAvailable`: PDF not available in PMC
- `PDFErrorDownloadFailed`: Download operation failed

**Example Error Handling:**
```go
err := pdfService.DownloadArticlePDF("12345", "test.pdf")
if err != nil {
    var pdfErr *PDFError
    if errors.As(err, &pdfErr) {
        switch pdfErr.Type {
        case PDFErrorArticleNotFound:
            fmt.Println("Article not found")
        case PDFErrorPMCIDNotFound:
            fmt.Println("Article not available in PMC")
        case PDFErrorPDFNotAvailable:
            fmt.Println("PDF not available")
        case PDFErrorDownloadFailed:
            fmt.Printf("Download failed: %v\n", pdfErr.Err)
        }
    }
}
```


## Thread Safety

The client is safe for concurrent use across multiple goroutines. All methods
can be called from different goroutines simultaneously.

## Rate Limiting

Please be mindful of NCBI's usage guidelines and rate limits. The library
currently does not implement automatic rate limiting, but future versions may
include this feature.



## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request


## Development

### Project Structure

```
literature/
├── doc.go              # Package documentation
├── literature.go       # Main client interface
├── types.go           # Public data structures
├── options.go         # Configuration options
├── errors.go          # Error types and handling
├── adapters.go        # Internal service adapters
├── internal/          # Private implementation details
├── examples/          # Usage examples
├── testdata/          # Test fixtures
└── cmd/pubmed/        # CLI tool (optional)
```

### Build and Format

```bash
# Format code
gofumpt -w .

# Lint codebase
golangcli-lint run

# Build
go build


```

### Testing

Quick test commands:

```bash
# Run all tests
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis

# Run specific test
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -run TestFetchArticle ./...

# Run with verbose output
gotestsum --format-hide-empty-pkg --format standard-verbose --format-icons hivis
```

For comprehensive testing guidelines, practices, and advanced usage, see [TESTING.md](TESTING.md).

## Support

For issues and questions, please use the GitHub issue tracker.
