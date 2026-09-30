/*
Package literature provides a clean, idiomatic Go client for accessing PubMed literature data.

The package offers a simple API for searching scientific articles, fetching metadata,
and accessing PDFs through the NCBI eUtils interface.

# Quick Start

	import "github.com/dictybase/literature"

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

# Configuration

The client can be configured with various options:

	client, err := literature.New(
		literature.WithTimeout(60*time.Second),
		literature.WithUserAgent("MyApp/1.0"),
		literature.WithHTTPClient(customHTTPClient),
	)

NCBI recommends including a `tool` name and developer `email` with every
E-utilities request. Supply your own API key; the package does not bundle one
or read environment variables:

	import "os"

	client, err := literature.New(
		literature.WithAPIKey(os.Getenv("NCBI_API_KEY")),
		literature.WithTool("my-tool"),
		literature.WithEmail(os.Getenv("NCBI_EMAIL")),
	)

`WithTool` rejects values containing spaces, and `WithEmail` validates the
address. An empty `NCBI_API_KEY` omits `api_key`, so NCBI's no-key limit applies.
See [NCBI API key instructions](https://www.ncbi.nlm.nih.gov/books/NBK25501/)
and [E-utilities usage policies](https://www.ncbi.nlm.nih.gov/books/NBK25497/).

# Error Handling

The package provides structured error types for better error handling:

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

# Search Operations

Search for articles using natural language queries:

	results, err := client.Search("cancer treatment",
		literature.WithLimit(20),
		literature.WithOffset(0),
	)

	for _, article := range results.Articles {
		fmt.Printf("Found: %s\n", article.Title)
	}

# Batch Operations

Fetch multiple articles efficiently:

	pmids := []string{"12345678", "87654321", "11111111"}
	articles, err := client.GetArticles(pmids)
	if err != nil {
		log.Fatal(err)
	}

	for _, article := range articles {
		fmt.Printf("Article: %s\n", article.Title)
	}

# PDF Access

Check for and retrieve PDF information:

	// Check if PDF is available
	available, err := client.HasPDF("12345678")
	if err != nil {
		log.Fatal(err)
	}

	if available {
		pdf, err := client.GetPDF("12345678")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("PDF URL: %s\n", pdf.URL)
	}

# Thread Safety

The client is safe for concurrent use across multiple goroutines.
All methods can be called from different goroutines simultaneously.

# Rate Limiting

NCBI's default E-utilities limits are 3 requests per second per IP address
without an API key and 10 requests per second with a key. Include `tool` and
`email` on every request. The package does not throttle PubMed requests; callers
must manage aggregate traffic across their processes and keep within NCBI's
limits. Each `Search` call makes two E-utilities requests. For higher rates,
request approval from NCBI. See the [E-utilities usage policies](https://www.ncbi.nlm.nih.gov/books/NBK25497/).

For more examples, see the examples/ directory in the source repository.
*/
package literature
