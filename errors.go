package literature

import "fmt"

// ErrorType represents the type of error that occurred.
type ErrorType string

const (
	// ErrorTypeInvalidInput indicates invalid input parameters.
	ErrorTypeInvalidInput ErrorType = "invalid_input"
	
	// ErrorTypeArticleNotFound indicates the requested article was not found.
	ErrorTypeArticleNotFound ErrorType = "article_not_found"
	
	// ErrorTypePDFNotFound indicates the requested PDF was not found.
	ErrorTypePDFNotFound ErrorType = "pdf_not_found"
	
	// ErrorTypePDFNotAvailable indicates the PDF is not available for this article.
	ErrorTypePDFNotAvailable ErrorType = "pdf_not_available"
	
	// ErrorTypeNetworkError indicates a network-related error.
	ErrorTypeNetworkError ErrorType = "network_error"
	
	// ErrorTypeParseError indicates an error parsing the API response.
	ErrorTypeParseError ErrorType = "parse_error"
	
	// ErrorTypeAPIError indicates an error from the PubMed API.
	ErrorTypeAPIError ErrorType = "api_error"
	
	// ErrorTypeTimeout indicates a request timeout.
	ErrorTypeTimeout ErrorType = "timeout"
	
	// ErrorTypeRateLimit indicates rate limiting by the API.
	ErrorTypeRateLimit ErrorType = "rate_limit"
)

// Error represents a literature client error with structured information.
type Error struct {
	Type     ErrorType `json:"type"`
	Message  string    `json:"message"`
	PMID     string    `json:"pmid,omitempty"`
	Query    string    `json:"query,omitempty"`
	Cause    error     `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.PMID != "" {
		return fmt.Sprintf("literature error [%s] for PMID %s: %s", e.Type, e.PMID, e.Message)
	}
	if e.Query != "" {
		return fmt.Sprintf("literature error [%s] for query '%s': %s", e.Type, e.Query, e.Message)
	}
	return fmt.Sprintf("literature error [%s]: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type.
func (e *Error) Is(target error) bool {
	if targetErr, ok := target.(*Error); ok {
		return e.Type == targetErr.Type
	}
	return false
}

// IsType checks if the error is of a specific ErrorType.
func (e *Error) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// NewError creates a new Error with the specified type and message.
func NewError(errorType ErrorType, message string) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
	}
}

// NewErrorWithPMID creates a new Error associated with a specific PMID.
func NewErrorWithPMID(errorType ErrorType, pmid, message string) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		PMID:    pmid,
	}
}

// NewErrorWithQuery creates a new Error associated with a specific query.
func NewErrorWithQuery(errorType ErrorType, query, message string) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Query:   query,
	}
}

// WrapError wraps an existing error with additional context.
func WrapError(errorType ErrorType, message string, cause error) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// WrapErrorWithPMID wraps an existing error with PMID context.
func WrapErrorWithPMID(errorType ErrorType, pmid, message string, cause error) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		PMID:    pmid,
		Cause:   cause,
	}
}

// Legacy error types for backward compatibility during transition
// TODO: Remove these once all internal services are migrated

// PDFErrorType represents different types of PDF download errors.
type PDFErrorType int

const (
	PDFErrorArticleNotFound PDFErrorType = iota
	PDFErrorPMCIDNotFound
	PDFErrorPDFNotAvailable
	PDFErrorDownloadFailed
)

// PDFError provides detailed error information for PDF download failures.
type PDFError struct {
	PMID string
	Type PDFErrorType
	Err  error
}

func (e *PDFError) Error() string {
	switch e.Type {
	case PDFErrorArticleNotFound:
		return fmt.Sprintf("article not found for PMID %s: %v", e.PMID, e.Err)
	case PDFErrorPMCIDNotFound:
		return fmt.Sprintf("no PMC ID found for PMID %s", e.PMID)
	case PDFErrorPDFNotAvailable:
		return fmt.Sprintf("PDF not available for PMID %s", e.PMID)
	case PDFErrorDownloadFailed:
		return fmt.Sprintf("download failed for PMID %s: %v", e.PMID, e.Err)
	default:
		return fmt.Sprintf("unknown error for PMID %s: %v", e.PMID, e.Err)
	}
}

func (e *PDFError) Unwrap() error {
	return e.Err
}
