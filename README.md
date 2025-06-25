# Literature - Go Client for PubMed

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

- [Modern API (Recommended)](#modern-api-recommended)
  - [Configuration](#configuration)
  - [Core Methods](#core-methods)
  - [Error Handling](#error-handling)
  - [Examples](#examples)
- [Legacy Services (Deprecated)](#legacy-services-deprecated)
  - [SearchService](#searchservice)
  - [ArticleService](#articleservice)
  - [PDFService](#pdfservice)
- [Types and Data Structures](#types-and-data-structures)
- [Project Structure](#project-structure)
- [Migration Guide](#migration-guide)

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

## Legacy Services (Deprecated)

⚠️ **Note**: The individual service classes below are deprecated. Use the unified `literature.Client` API above for new projects.

## Services

### SearchService

Handles PubMed search operations using the E-utilities search API.

#### Constructor

```go
func NewSearchService(options ...SearchServiceOption) *SearchService
```

#### Options

- `WithSearchHTTPClient(client *http.Client)`: Set custom HTTP client
- `WithRetmax(retmax int)`: Set maximum number of results (default: 10)

#### Methods

##### SearchPubMed

```go
func (s *SearchService) SearchPubMed(query string) (*ESearchResult, error)
```

Performs a search query against PubMed and returns search results.

**Parameters:**
- `query`: Search query string (e.g., "diabetes AND treatment")

**Returns:**
- `*ESearchResult`: Search results containing IDs and metadata
- `error`: Error if search fails

**Example:**
```go
searchService := NewSearchService(WithRetmax(20))
result, err := searchService.SearchPubMed("machine learning in healthcare")
if err != nil {
    log.Fatal(err)
}

pmids := result.GetIDs()
fmt.Printf("Found %d articles\n", len(pmids))
```

##### FetchPubMedDetails

```go
func (s *SearchService) FetchPubMedDetails(webEnv, queryKey string) (*PubMedArticleSet, error)
```

Retrieves detailed article information using WebEnv and QueryKey from search results.

**Parameters:**
- `webEnv`: Web environment string from search results
- `queryKey`: Query key from search results

**Returns:**
- `*PubMedArticleSet`: Set of detailed article metadata
- `error`: Error if fetch fails

**Example:**
```go
searchResult, _ := searchService.SearchPubMed("COVID-19")
articles, err := searchService.FetchPubMedDetails(
    searchResult.WebEnv,
    searchResult.QueryKey,
)
if err != nil {
    log.Fatal(err)
}

for _, article := range articles.PubMedArticles {
    fmt.Printf("Title: %s\n", article.GetTitle())
}
```

### ArticleService

Handles fetching PubMed article metadata for specific PMIDs.

#### Constructor

```go
func NewArticleService() *ArticleService
```

#### Methods

##### FetchArticle

```go
func (s *ArticleService) FetchArticle(pmid string) (*PubMedArticle, error)
```

Retrieves article metadata for the given PMID.

**Parameters:**
- `pmid`: PubMed ID as string

**Returns:**
- `*PubMedArticle`: Complete article metadata
- `error`: Error if article not found or fetch fails

**Example:**
```go
articleService := NewArticleService()
article, err := articleService.FetchArticle("33515252")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", article.GetTitle())
fmt.Printf("Journal: %s\n", article.GetJournalTitle())
fmt.Printf("Year: %s\n", article.GetPubYear())
fmt.Printf("Abstract: %s\n", article.GetAbstract())

// Get authors
authors := article.GetAuthors()
for _, author := range authors {
    fmt.Printf("Author: %s %s\n", author.ForeName, author.LastName)
}

// Get DOI if available
if doi, found := article.GetDOI(); found {
    fmt.Printf("DOI: %s\n", doi)
}
```

### PDFService

Handles PDF link discovery and downloading from PMC (PubMed Central).

#### Constructor

```go
func NewPDFService(options ...PDFServiceOption) *PDFService
```

#### Options

- `WithHTTPClient(client *http.Client)`: Set custom HTTP client

#### Methods

##### IsPDFAvailable

```go
func (s *PDFService) IsPDFAvailable(pmid string) (bool, error)
```

Checks if a PDF is available for the given PMID and caches download info. Must
be called before `DownloadPDF`.

**Parameters:**
- `pmid`: PubMed ID as string

**Returns:**
- `bool`: True if PDF is available
- `error`: Error if check fails

##### GetPDFURL

```go
func (s *PDFService) GetPDFURL() (string, error)
```

Returns the direct download URL using cached download info.
`IsPDFAvailable` must be called first and return true.

**Returns:**
- `string`: Direct download URL for the PDF
- `error`: Error if no cached download info available

##### DownloadPDF

```go
func (s *PDFService) DownloadPDF(filePath string) error
```

Downloads the PDF using cached download info to the specified file.
`IsPDFAvailable` must be called first and return true.

**Parameters:**
- `filePath`: Local file path where PDF should be saved

**Returns:**
- `error`: Error if download fails

##### DownloadArticlePDF (Convenience Method)

```go
func (s *PDFService) DownloadArticlePDF(pmid, filePath string) error
```

Convenience method that combines availability check and downloading.

**Parameters:**
- `pmid`: PubMed ID as string
- `filePath`: Local file path where PDF should be saved

**Returns:**
- `error`: Error if PDF not available or download fails

**Example:**
```go
pdfService := NewPDFService()

// Method 1: Check availability first, then get URL
available, err := pdfService.IsPDFAvailable("33515252")
if err != nil {
    log.Fatal(err)
}

if available {
    url, err := pdfService.GetPDFURL()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("PDF URL: %s\n", url)
}

// Method 2: Check availability first, then download
available, err = pdfService.IsPDFAvailable("33515252")
if err != nil {
    log.Fatal(err)
}

if available {
    err = pdfService.DownloadPDF("article.pdf")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("PDF downloaded successfully")
}

// Method 3: Use convenience method
err = pdfService.DownloadArticlePDF("33515252", "article.pdf")
if err != nil {
    var pdfErr *PDFError
    if errors.As(err, &pdfErr) {
        switch pdfErr.Type {
        case PDFErrorPDFNotAvailable:
            fmt.Println("PDF not available for this article")
        case PDFErrorDownloadFailed:
            fmt.Printf("Download failed: %v\n", err)
        }
    }
}
```

##### GetCurrentPMID

```go
func (s *PDFService) GetCurrentPMID() string
```

Returns the currently cached PMID, empty string if none.

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

## Project Structure

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

## Migration Guide

### From Legacy Services to Modern API

If you're using the old service-based approach, here's how to migrate:

#### Before (Legacy)
```go
// Old way - multiple services
articleService := NewArticleService()
searchService := NewSearchService()
pdfService := NewPDFService()

// Fetch article
article, err := articleService.FetchArticle("12345678")

// Search
results, err := searchService.SearchPubMed("query")

// PDF operations
available, err := pdfService.IsPDFAvailable("12345678")
```

#### After (Modern)
```go
// New way - unified client
client, err := literature.New()

// Fetch article
article, err := client.GetArticle("12345678")

// Search
results, err := client.Search("query")

// PDF operations
available, err := client.HasPDF("12345678")
```

### Benefits of Migration

1. **Simplified API** - Single client instead of multiple services
2. **Better Error Handling** - Structured error types with context
3. **Configuration** - Centralized configuration with options pattern
4. **Thread Safety** - Safe for concurrent use
5. **Future-Proof** - Active development and new features

## Thread Safety

The client is safe for concurrent use across multiple goroutines. All methods can be called from different goroutines simultaneously.

## Rate Limiting

Please be mindful of NCBI's usage guidelines and rate limits. The library currently does not implement automatic rate limiting, but future versions may include this feature.

## Utility Functions

The library includes generic utility functions for functional programming operations:

### Find

```go
func Find[T any](slice []T, predicate func(T) bool) (*T, bool)
```

Finds the first element in a slice matching the predicate.

### Map

```go
func Map[T, U any](ts []T, f func(T) U) []U
```

Transforms a slice using the provided function.

### Filter

```go
func Filter[T any](slice []T, predicate func(T) bool) []T
```

Filters a slice using the provided predicate.

## Usage Examples

### Complete Workflow Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // 1. Search for articles
    searchService := NewSearchService(WithRetmax(5))
    searchResult, err := searchService.SearchPubMed("machine learning healthcare")
    if err != nil {
        log.Fatal(err)
    }

    pmids := searchResult.GetIDs()
    fmt.Printf("Found %d articles\n", len(pmids))

    // 2. Get detailed article information
    articleService := NewArticleService()
    pdfService := NewPDFService()

    for _, pmid := range pmids {
        // Fetch article metadata
        article, err := articleService.FetchArticle(pmid)
        if err != nil {
            fmt.Printf("Error fetching article %s: %v\n", pmid, err)
            continue
        }

        fmt.Printf("\nPMID: %s\n", article.GetPMID())
        fmt.Printf("Title: %s\n", article.GetTitle())
        fmt.Printf("Journal: %s (%s)\n", article.GetJournalTitle(), article.GetPubYear())
        
        // Check for DOI
        if doi, found := article.GetDOI(); found {
            fmt.Printf("DOI: %s\n", doi)
        }

        // Check PDF availability and get URL if available
        available, err := pdfService.IsPDFAvailable(pmid)
        if err != nil {
            fmt.Printf("Error checking PDF availability: %v\n", err)
            continue
        }

        if available {
            pdfURL, err := pdfService.GetPDFURL()
            if err != nil {
                fmt.Printf("Error getting PDF URL: %v\n", err)
                continue
            }
            fmt.Printf("PDF URL: %s\n", pdfURL)
            
            // Optionally download the PDF
            filename := fmt.Sprintf("%s.pdf", pmid)
            err = pdfService.DownloadPDF(filename)
            if err == nil {
                fmt.Printf("PDF downloaded: %s\n", filename)
            }
        } else {
            fmt.Println("PDF not available")
        }
    }
}
```

### Advanced Search with Batch Processing

```go
func batchProcessArticles(query string, batchSize int) error {
    searchService := NewSearchService(WithRetmax(batchSize))
    result, err := searchService.SearchPubMed(query)
    if err != nil {
        return err
    }

    // Get detailed articles using WebEnv/QueryKey for efficiency
    articles, err := searchService.FetchPubMedDetails(result.WebEnv, result.QueryKey)
    if err != nil {
        return err
    }

    pdfService := NewPDFService()
    
    for _, article := range articles.PubMedArticles {
        fmt.Printf("Processing: %s\n", article.GetTitle())
        
        // Extract author names
        authors := article.GetAuthors()
        authorNames := Map(authors, func(a Author) string {
            return fmt.Sprintf("%s %s", a.ForeName, a.LastName)
        })
        
        fmt.Printf("Authors: %v\n", authorNames)
        
        // Attempt PDF download
        pmid := article.GetPMID()
        err := pdfService.DownloadArticlePDF(pmid, pmid+".pdf")
        if err == nil {
            fmt.Printf("Downloaded PDF for %s\n", pmid)
        }
    }
    
    return nil
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Testing

Run tests with the provided commands:

```bash
# Run all tests
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis

# Run specific test
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -run TestFetchArticle ./...

# Run with verbose output
gotestsum --format-hide-empty-pkg --format standard-verbose --format-icons hivis
```

## Development

### Build and Format

```bash
# Format code
gofumpt -w .

# Lint codebase
golangcli-lint run

# Build
go build
```

## License

[Your chosen license]

## Support

For issues and questions, please use the GitHub issue tracker.
