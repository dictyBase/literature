# Literature - Go Client for fetching literature information

[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](https://github.com/dictyBase/literature/blob/develop/LICENSE)
![GitHub action](https://github.com/dictyBase/literature/workflows/Continuous%20integration/badge.svg)
[![codecov](https://codecov.io/gh/dictyBase/literature/graph/badge.svg?token=RE3OI8BCJS)](https://codecov.io/gh/dictyBase/literature)
![Last commit](https://badgen.net/github/last-commit/dictyBase/literature/develop)
[![Funding](https://badgen.net/badge/Funding/Rex%20L%20Chisholm,dictyBase,DCR/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=10024726&icde=0)

A production-ready, idiomatic Go library for accessing scientific literature
through multiple APIs. Supports both PubMed (NCBI eUtils) for authoritative
biomedical research and EuropePMC for enhanced European content with rich
metadata and citation analytics.

## Features

### Core Literature Access
- 🔍 **Advanced Search**: Comprehensive literature search with natural language and structured queries
- 📄 **Article Retrieval**: Fetch detailed article metadata by PMID with rich information
- 📚 **Batch Operations**: Efficient processing of multiple articles simultaneously
- 🔗 **Full-Text Access**: PDF and full-text availability checking and URL retrieval
- 🔄 **Related Articles**: Discover similar articles and citations

### Technical Features
- ⚙️ **Configurable Clients**: Custom HTTP clients with timeout and retry policies
- 🛡️ **Robust Error Handling**: Structured error types with detailed context
- 🔄 **Thread-Safe**: Safe for concurrent use across multiple goroutines
- 🔧 **Functional Options**: Clean configuration using the options pattern
- 📖 **Comprehensive Documentation**: Extensive examples and usage guides
- 🚀 **High Performance**: Optimized for concurrent operations

### API-Specific Features

#### PubMed (NCBI eUtils) - Distinctive Features
- 🏛️ **NCBI Integration**: Direct access to the authoritative NCBI PubMed database
- 🔬 **Biomedical Focus**: Optimized for life sciences and biomedical research
- 📊 **MeSH Terms**: Access to Medical Subject Headings for precise categorization
- 🌐 **Global Coverage**: Comprehensive international biomedical literature

#### EuropePMC - Distinctive Features
- 🇪🇺 **European Excellence**: Enhanced coverage of European research and journals
- 👥 **Rich Author Data**: Detailed author information with ORCID IDs and institutional affiliations
- 📊 **Citation Analytics**: Real-time citation counts and impact metrics
- 🔗 **Multiple Formats**: Access to full-text in PDF, HTML, and XML formats
- 💰 **Funding Information**: Comprehensive grant and funding agency details
- 🆓 **Open Access Focus**: Enhanced open access detection and licensing information
- 🌍 **Multi-Language Support**: Better support for non-English European literature

## Installation

```bash
go get github.com/dictybase/literature
```

**Requirements:**
- Go 1.23.8 or later

## Quick Start

Choose the API that best fits your research needs:

## API Comparison

| Feature | PubMed (NCBI eUtils) | EuropePMC |
|---------|---------------------|------------|
| **Data Source** | NCBI PubMed | European PMC Database |
| **Coverage** | 35M+ citations | 37M+ citations |
| **Best For** | Biomedical research | European research + broader coverage |
| **Author Info** | Basic | Rich (ORCID, affiliations) |
| **Citation Metrics** | No | Yes (real-time counts) |
| **Full Text Access** | PDF | Enhanced PDF/HTML/XML |
| **Funding Data** | No | Yes (grants, agencies) |
| **Open Access** | Basic detection | Enhanced licensing info |
| **Rate Limits** | 3 req/sec | 10 req/sec (configurable) |
| **Built-in Rate Limiting** | No | Yes |
| **Retry Logic** | No | Yes |

### PubMed (NCBI eUtils) 

```go
package main

import (
    "fmt"
    "log"

    "github.com/dictybase/literature"
)

func main() {
    // Create a PubMed client for biomedical literature
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

### EuropePMC 

```go
package main

import (
    "fmt"
    "log"

    "github.com/dictybase/literature"
)

func main() {
    // Create an EuropePMC client for rich metadata and European content
    client, err := literature.NewEuropePMCClient()
    if err != nil {
        log.Fatal(err)
    }

    // Fetch an article with enhanced metadata
    article, err := client.GetArticle("12345678")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", article.Title)
    fmt.Printf("Journal: %s\n", article.Journal.Title)
    fmt.Printf("Citations: %d\n", article.CitedByCount)
    fmt.Printf("Open Access: %t\n", article.IsOpenAccess)
    
    // Show author affiliations (EuropePMC-specific feature)
    for _, author := range article.Authors {
        fmt.Printf("Author: %s\n", author.FullName)
        for _, affiliation := range author.Affiliations {
            fmt.Printf("  - %s\n", affiliation.Affiliation)
        }
    }
}
```

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Comparison](#api-comparison)
- [PubMed API (NCBI eUtils)](#pubmed-api-ncbi-eutils)
  - [Configuration](#configuration)
  - [Core Methods](#core-methods)
  - [Error Handling](#error-handling)
  - [Examples](#examples)
- [EuropePMC API](#europepmc-api)
  - [Configuration](#europepmc-configuration)
  - [Core Methods](#europepmc-core-methods)
  - [Search Options](#europepmc-search-options)
  - [Examples](#europepmc-examples)
- [Types and Data Structures](#types-and-data-structures)
- [Error Handling](#error-handling-1)
- [Thread Safety](#thread-safety)
- [Rate Limiting](#rate-limiting)
- [Performance](#performance)
- [Contributing](#contributing)
- [Development](#development)
  - [Project Structure](#project-structure)
- [Support](#support)

## PubMed API (NCBI eUtils)

The `literature` package provides a clean, unified interface for all PubMed operations.

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
- `DownloadPDF(pmid, filePath string) error` - Downloads the PDF for a given PMID to the specified path.

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

- [`examples/basic/`](examples/basic/) - Basic PubMed usage patterns
- [`examples/advanced/`](examples/advanced/) - Advanced PubMed configuration and error handling
- [`examples/europepmc/`](examples/europepmc/) - EuropePMC usage examples

## EuropePMC API

The EuropePMC client provides access to the comprehensive European literature database with enhanced metadata and European research content.

### EuropePMC Configuration

Customize the EuropePMC client with various options:

```go
client, err := literature.NewEuropePMCClient(
    literature.WithEuropePMCTimeout(60*time.Second),
    literature.WithEuropePMCUserAgent("MyApp/1.0"),
    literature.WithEuropePMCEmail("contact@myapp.com"),
    literature.WithEuropePMCRetryPolicy(5, 2*time.Second),
    literature.WithEuropePMCRateLimit(5.0), // 5 requests per second
)
```

### EuropePMC Core Methods

- `GetArticle(pmid string) (*EuropePMCArticle, error)` - Fetch single article with comprehensive metadata
- `GetArticles(pmids []string) ([]*EuropePMCArticle, error)` - Fetch multiple articles efficiently
- `Search(query string, opts ...EuropePMCSearchOption) (*EuropePMCSearchResult, error)` - Advanced literature search
- `FindSimilar(pmid string, opts ...EuropePMCSearchOption) (*EuropePMCSearchResult, error)` - Find related articles
- `HasPDF(pmid string) (bool, error)` - Check PDF availability
- `GetPDFURLs(pmid string) ([]EuropePMCFullTextURL, error)` - Get all available PDF URLs

#### EuropePMC Configuration Options

- `WithEuropePMCHTTPClient(client *http.Client)` - Custom HTTP client
- `WithEuropePMCTimeout(timeout time.Duration)` - Request timeout
- `WithEuropePMCBaseURL(url string)` - Custom API base URL (for testing)
- `WithEuropePMCUserAgent(userAgent string)` - Custom User-Agent header
- `WithEuropePMCEmail(email string)` - Contact email for high-volume usage
- `WithEuropePMCRetryPolicy(maxRetries int, retryDelay time.Duration)` - Retry configuration
- `WithEuropePMCRateLimit(requestsPerSecond float64)` - Rate limiting
- `WithEuropePMCDefaultResultType(resultType string)` - Default result type ("core", "lite")
- `WithEuropePMCDefaultFormat(format string)` - Default response format ("json", "xml")

### EuropePMC Search Options

- `WithEuropePMCLimit(limit int)` - Maximum results per search (default: 20)
- `WithEuropePMCOffset(offset int)` - Pagination offset
- `WithEuropePMCResultType(resultType string)` - Result detail level ("core", "lite")
- `WithEuropePMCFormat(format string)` - Response format ("json", "xml")

### EuropePMC Examples

#### Basic Article Retrieval

```go
client, err := literature.NewEuropePMCClient()
if err != nil {
    log.Fatal(err)
}

// Get detailed article information
article, err := client.GetArticle("25844567")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", article.Title)
fmt.Printf("Journal: %s (%s)\n", article.Journal.Title, article.Journal.ISOAbbreviation)
fmt.Printf("Publication Year: %s\n", article.PubYear)
fmt.Printf("Open Access: %t\n", article.IsOpenAccess)
fmt.Printf("Citations: %d\n", article.CitedByCount)

// Print authors with affiliations
for _, author := range article.Authors {
    fmt.Printf("Author: %s", author.FullName)
    if author.ORCID != "" {
        fmt.Printf(" (ORCID: %s)", author.ORCID)
    }
    fmt.Println()
    
    for _, affiliation := range author.Affiliations {
        fmt.Printf("  - %s\n", affiliation.Affiliation)
    }
}

// Print MeSH headings
if len(article.MeshHeadings) > 0 {
    fmt.Println("\nMeSH Headings:")
    for _, mesh := range article.MeshHeadings {
        majorTopic := ""
        if mesh.MajorTopic {
            majorTopic = " (Major Topic)"
        }
        fmt.Printf("  - %s%s\n", mesh.DescriptorName, majorTopic)
    }
}
```

#### Advanced Search with Options

```go
client, err := literature.NewEuropePMCClient(
    literature.WithEuropePMCEmail("research@university.edu"),
)
if err != nil {
    log.Fatal(err)
}

// Search for articles with advanced options
searchResult, err := client.Search(
    "cancer AND immunotherapy",
    literature.WithEuropePMCLimit(50),
    literature.WithEuropePMCResultType("core"),
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d articles for query: %s\n", searchResult.Total, searchResult.Query)

for i, article := range searchResult.Articles {
    fmt.Printf("%d. %s\n", i+1, article.Title)
    fmt.Printf("   Journal: %s (%s)\n", article.Journal.Title, article.PubYear)
    fmt.Printf("   Open Access: %t, Has PDF: %t\n", article.IsOpenAccess, article.HasPDF)
    
    if article.Abstract != "" {
        abstract := article.Abstract
        if len(abstract) > 200 {
            abstract = abstract[:200] + "..."
        }
        fmt.Printf("   Abstract: %s\n", abstract)
    }
    fmt.Println()
}
```

#### PDF Access and Full Text URLs

```go
client, err := literature.NewEuropePMCClient()
if err != nil {
    log.Fatal(err)
}

pmid := "25844567"

// Check if PDF is available
hasPDF, err := client.HasPDF(pmid)
if err != nil {
    log.Fatal(err)
}

if hasPDF {
    // Get all available PDF URLs
    pdfURLs, err := client.GetPDFURLs(pmid)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Available PDF sources for PMID %s:\n", pmid)
    for _, pdfURL := range pdfURLs {
        fmt.Printf("  - %s: %s\n", pdfURL.Site, pdfURL.URL)
        fmt.Printf("    Availability: %s (%s)\n", 
                   pdfURL.Availability, pdfURL.AvailabilityCode)
    }
} else {
    fmt.Printf("No PDF available for PMID %s\n", pmid)
}
```

#### Finding Similar Articles

```go
client, err := literature.NewEuropePMCClient()
if err != nil {
    log.Fatal(err)
}

// Find articles similar to a given PMID
similarResult, err := client.FindSimilar(
    "25844567",
    literature.WithEuropePMCLimit(10),
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d similar articles:\n", len(similarResult.Articles))
for _, article := range similarResult.Articles {
    fmt.Printf("- %s (%s)\n", article.Title, article.PubYear)
    fmt.Printf("  PMID: %s, Citations: %d\n", article.PMID, article.CitedByCount)
}
```

#### Error Handling with EuropePMC

```go
article, err := client.GetArticle("invalid-pmid")
if err != nil {
    if litErr, ok := err.(*literature.Error); ok {
        switch litErr.Type {
        case literature.ErrorTypeInvalidInput:
            fmt.Println("Invalid PMID format")
        case literature.ErrorTypeArticleNotFound:
            fmt.Println("Article not found in EuropePMC")
        case literature.ErrorTypeAPIError:
            fmt.Printf("EuropePMC API error: %s\n", litErr.Message)
        case literature.ErrorTypeNetworkError:
            fmt.Println("Network error - check connection")
        }
    }
}
```

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

### EuropePMC Types

#### EuropePMCArticle

Represents a comprehensive article from EuropePMC with rich metadata.

```go
type EuropePMCArticle struct {
    ID           string                 `json:"id"`
    Source       string                 `json:"source"`
    PMID         string                 `json:"pmid"`
    PMCID        string                 `json:"pmcid,omitempty"`
    DOI          string                 `json:"doi,omitempty"`
    Title        string                 `json:"title"`
    AuthorString string                 `json:"author_string"`
    Authors      []EuropePMCAuthor      `json:"authors"`
    Abstract     string                 `json:"abstract"`
    Journal      EuropePMCJournal       `json:"journal"`
    PubYear      string                 `json:"pub_year"`
    PageInfo     string                 `json:"page_info,omitempty"`
    Keywords     []string               `json:"keywords,omitempty"`
    FullTextURLs []EuropePMCFullTextURL `json:"full_text_urls,omitempty"`
    IsOpenAccess bool                   `json:"is_open_access"`
    HasPDF       bool                   `json:"has_pdf"`
    License      string                 `json:"license,omitempty"`
    CitedByCount int                    `json:"cited_by_count"`
    Language     string                 `json:"language,omitempty"`
    PubTypes     []string               `json:"pub_types,omitempty"`
    MeshHeadings []EuropePMCMeshHeading `json:"mesh_headings,omitempty"`
    Chemicals    []EuropePMCChemical    `json:"chemicals,omitempty"`
    Grants       []EuropePMCGrant       `json:"grants,omitempty"`
    PublishDate  *time.Time             `json:"publish_date,omitempty"`
    CreationDate *time.Time             `json:"creation_date,omitempty"`
    RevisionDate *time.Time             `json:"revision_date,omitempty"`
}
```

#### EuropePMCAuthor

```go
type EuropePMCAuthor struct {
    FullName     string                       `json:"full_name"`
    FirstName    string                       `json:"first_name"`
    LastName     string                       `json:"last_name"`
    Initials     string                       `json:"initials"`
    ORCID        string                       `json:"orcid,omitempty"`
    Affiliations []EuropePMCAuthorAffiliation `json:"affiliations,omitempty"`
}
```

#### EuropePMCJournal

```go
type EuropePMCJournal struct {
    Title               string `json:"title"`
    MedlineAbbreviation string `json:"medline_abbreviation,omitempty"`
    ISOAbbreviation     string `json:"iso_abbreviation,omitempty"`
    ISSN                string `json:"issn,omitempty"`
    ESSN                string `json:"essn,omitempty"`
    Volume              string `json:"volume,omitempty"`
    Issue               string `json:"issue,omitempty"`
    IssueID             int    `json:"issue_id,omitempty"`
    DateOfPublication   string `json:"date_of_publication,omitempty"`
    MonthOfPublication  int    `json:"month_of_publication,omitempty"`
    YearOfPublication   int    `json:"year_of_publication,omitempty"`
    NLMID               string `json:"nlm_id,omitempty"`
}
```

#### EuropePMCSearchResult

```go
type EuropePMCSearchResult struct {
    Query    string              `json:"query"`
    Total    int                 `json:"total"`
    Articles []*EuropePMCArticle `json:"articles"`
    Limit    int                 `json:"limit"`
    Offset   int                 `json:"offset"`
}
```

#### EuropePMCFullTextURL

```go
type EuropePMCFullTextURL struct {
    Availability     string `json:"availability"`
    AvailabilityCode string `json:"availability_code"`
    DocumentStyle    string `json:"document_style"`
    Site             string `json:"site"`
    URL              string `json:"url"`
}
```

#### EuropePMCMeshHeading

```go
type EuropePMCMeshHeading struct {
    MajorTopic     bool                     `json:"major_topic"`
    DescriptorName string                   `json:"descriptor_name"`
    MeshQualifiers []EuropePMCMeshQualifier `json:"mesh_qualifiers,omitempty"`
}
```

#### EuropePMCGrant

```go
type EuropePMCGrant struct {
    GrantID string `json:"grant_id"`
    Agency  string `json:"agency"`
    OrderIn int    `json:"order_in"`
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

Both clients are safe for concurrent use across multiple goroutines. All methods can be called from different goroutines simultaneously.

```go
// Safe to use from multiple goroutines
var wg sync.WaitGroup
pmids := []string{"12345678", "87654321", "11111111"}

for _, pmid := range pmids {
    wg.Add(1)
    go func(pmid string) {
        defer wg.Done()
        article, err := client.GetArticle(pmid)
        if err != nil {
            log.Printf("Error fetching %s: %v", pmid, err)
            return
        }
        fmt.Printf("Fetched: %s\n", article.Title)
    }(pmid)
}

wg.Wait()
```

## Rate Limiting

### PubMed Rate Limits
NCBI requests that users:
- Make no more than 3 requests per second for E-utilities
- Use the `tool` and `email` parameters for identification
- Consider using the History Server for large batch operations

### EuropePMC Rate Limits
The EuropePMC client includes built-in rate limiting:
- Default: 10 requests per second
- Configurable via `WithEuropePMCRateLimit(float64)`
- Automatically handles retry with exponential backoff

```go
// Configure rate limiting for EuropePMC
client, err := literature.NewEuropePMCClient(
    literature.WithEuropePMCRateLimit(5.0), // 5 requests per second
    literature.WithEuropePMCRetryPolicy(5, 2*time.Second),
)
```

**Best Practices:**
- Use batch operations (`GetArticles`) when fetching multiple articles
- Cache results when possible to reduce API calls
- Monitor rate limit headers in API responses
- Implement exponential backoff for failed requests

## Performance

### Benchmarks

The library is optimized for performance with the following characteristics:

- **Single Article Fetch**: ~200-500ms (network dependent)
- **Batch Operations**: ~50-100ms per article in batch
- **Memory Usage**: ~1-5MB per 1000 articles (depending on metadata richness)
- **Concurrent Safety**: Lock-free operations, scales with goroutines

### Optimization Tips

```go
// Prefer batch operations for multiple articles
// Efficient: Single API call
articles, err := client.GetArticles([]string{"12345678", "87654321", "11111111"})

// Less efficient: Multiple API calls
for _, pmid := range pmids {
    article, err := client.GetArticle(pmid)
    // ...
}

// Use context with timeout for long-running operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// EuropePMC: Configure appropriate rate limiting
client, err := literature.NewEuropePMCClient(
    literature.WithEuropePMCRateLimit(5.0), // Adjust based on usage patterns
    literature.WithEuropePMCRetryPolicy(3, time.Second),
)
```

### Memory Management

- Articles are processed in streaming fashion when possible
- Large result sets are paginated automatically
- Consider implementing result caching for frequently accessed articles
- Use `defer` statements to ensure proper cleanup of HTTP connections



## Contributing

We welcome contributions! Please follow these guidelines:

### Development Workflow

1. **Fork and Clone**
   ```bash
   git clone https://github.com/yourusername/literature.git
   cd literature
   ```

2. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Follow Code Standards**
   ```bash
   # Format code
   gofumpt -w .
   
   # Run linter
   golangci-lint run
   
   # Run tests
   gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis
   ```

4. **Add Comprehensive Tests**
   - Unit tests for all new functionality
   - Integration tests for API interactions
   - Error path testing
   - Documentation examples that compile

5. **Update Documentation**
   - Update README.md if adding new features
   - Add or update Go doc comments
   - Include usage examples

6. **Submit Pull Request**
   - Ensure all CI checks pass
   - Include a clear description of changes
   - Reference any related issues

### Code Style Guidelines

See [CLAUDE.md](CLAUDE.md) for detailed coding conventions including:
- Go coding standards and best practices
- Error handling patterns with structured error types
- Testing methodologies using gotestsum
- Documentation requirements and examples
- Functional programming utilities usage
- Options pattern implementation
- Validation with go-playground/validator

### Issue Templates

When reporting bugs or requesting features, please include:

**For Bugs:**
- Go version and OS
- Library version
- Minimal reproduction code
- Expected vs actual behavior
- API response samples (with sensitive data removed)

**For Features:**
- Use case description
- Proposed API design
- Compatibility considerations
- Performance implications


## Development

### Project Structure

```
literature/
├── doc.go              # Package documentation
├── literature.go       # Main PubMed client interface
├── europepmc.go        # EuropePMC client interface
├── types.go           # PubMed data structures
├── europepmc_types.go  # EuropePMC data structures
├── options.go         # PubMed configuration options
├── europepmc_options.go # EuropePMC configuration options
├── errors.go          # Error types and handling
├── adapters.go        # Internal service adapters
├── internal/          # Private implementation details
│   ├── pubmed_service.go    # PubMed API client
│   └── europepmc_service.go # EuropePMC API client
├── examples/          # Usage examples
│   ├── basic/         # Basic PubMed examples
│   ├── advanced/      # Advanced PubMed examples
│   └── europepmc/     # EuropePMC examples
├── testdata/          # Test fixtures
└── cmd/pubmed/        # CLI tool (optional)
```

### Build and Format

```bash
# Format code
gofumpt -w .

# Lint codebase
golangci-lint run

# Build
go build

# Run benchmarks
go test -bench=. -benchmem

# Check for race conditions
go test -race ./...
```

### Testing

Quick test commands:

```bash
# Run all tests
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis

# Run specific test
gotestsum --format-hide-empty-pkg --format testdox --format-icons hivis -- -run TestFindSimilar ./...

# Run with verbose output
gotestsum --format-hide-empty-pkg --format standard-verbose --format-icons hivis

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Integration tests (requires network)
go test -tags=integration ./...
```

For comprehensive testing guidelines, practices, and advanced usage, see [TESTING.md](TESTING.md).

## Support

For issues and questions, please use the [GitHub issue tracker](https://github.com/dictyBase/literature/issues).

- 🐛 **Bug Reports**: Found a bug? [Create an issue](https://github.com/dictyBase/literature/issues/new)
- 💡 **Feature Requests**: Have an idea? We'd love to hear it!
- ❓ **Questions**: Need help? Check existing issues or create a new one
- 📖 **Documentation**: Improvements to docs are always welcome
