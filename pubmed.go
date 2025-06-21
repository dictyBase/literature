package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// OAResponse represents the top-level XML structure for the OA FTP service.
type OAResponse struct {
	XMLName xml.Name `xml:"OA"`
	Records struct {
		Records []OARecord `xml:"record"`
	} `xml:"records"`
}

// OARecord contains information about a specific record, including download links.
type OARecord struct {
	XMLName xml.Name `xml:"record"`
	ID      string   `xml:"id,attr"`
	Links   []OALink `xml:"link"`
}

// OALink represents a download link for a specific format (e.g., pdf, tgz).
type OALink struct {
	XMLName xml.Name `xml:"link"`
	Format  string   `xml:"format,attr"`
	HREF    string   `xml:"href,attr"`
}

// ESearchResult represents the top-level XML structure for esearch.
type ESearchResult struct {
	XMLName  xml.Name `xml:"eSearchResult"`
	Count    string   `xml:"Count"`
	RetMax   string   `xml:"RetMax"`
	RetStart string   `xml:"RetStart"`
	QueryKey string   `xml:"QueryKey"`
	WebEnv   string   `xml:"WebEnv"`
	IDList   struct {
		IDs []string `xml:"Id"`
	} `xml:"IdList"`
}

// PubMedArticleSet represents the top-level XML structure for efetch.
type PubMedArticleSet struct {
	XMLName        xml.Name        `xml:"PubmedArticleSet"`
	PubMedArticles []PubMedArticle `xml:"PubmedArticle"`
}

// Author represents an author of an article.
type Author struct {
	LastName string `xml:"LastName"`
	ForeName string `xml:"ForeName"`
}

// ArticleID represents an identifier for an article, like DOI or PMID.
type ArticleID struct {
	IDType string `xml:"IdType,attr"`
	Value  string `xml:",chardata"`
}

// PubMedArticle represents a single PubMed article.
type PubMedArticle struct {
	XMLName         xml.Name `xml:"PubmedArticle"`
	MedlineCitation struct {
		Article struct {
			Journal struct {
				Title        string `xml:"Title"`
				JournalIssue struct {
					PubDate struct {
						Year  string `xml:"Year"`
						Month string `xml:"Month"`
					} `xml:"PubDate"`
				} `xml:"JournalIssue"`
			} `xml:"Journal"`
			ArticleTitle string `xml:"ArticleTitle"`
			Pagination   struct {
				MedlinePgn string `xml:"MedlinePgn"`
			} `xml:"Pagination"`
			Abstract struct {
				AbstractText string `xml:"AbstractText"`
			} `xml:"Abstract"`
			AuthorList struct {
				Authors []Author `xml:"Author"`
			} `xml:"AuthorList"`
		} `xml:"Article"`
		PMID string `xml:"PMID"`
	} `xml:"MedlineCitation"`
	PubmedData struct {
		ArticleIdList struct {
			ArticleIds []ArticleID `xml:"ArticleId"`
		} `xml:"ArticleIdList"`
	} `xml:"PubmedData"`
}

func searchPubMed(query string) (*ESearchResult, error) {
	esearchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=%s&retmax=10&retmode=xml&usehistory=y",
		url.QueryEscape(query),
	)

	// #nosec G107
	resp, err := http.Get(esearchURL)
	if err != nil {
		return nil, fmt.Errorf("error making esearch request: %w", err)
	}
	defer resp.Body.Close()

	esearchResult := &ESearchResult{}
	if err := xml.NewDecoder(resp.Body).Decode(esearchResult); err != nil {
		return nil, fmt.Errorf("error unmarshaling esearch XML: %w", err)
	}

	return esearchResult, nil
}

func fetchPubMedDetails(webEnv, queryKey string) (*PubMedArticleSet, error) {
	efetchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&retmode=xml&WebEnv=%s&query_key=%s&retmax=10",
		webEnv,
		queryKey,
	)

	// #nosec G107
	resp, err := http.Get(efetchURL)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", err)
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling efetch XML: %w", err)
	}

	return articleSet, nil
}

func fetchPubMedArticle(pmid string) (*PubMedArticleSet, error) {
	efetchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&retmode=xml&id=%s",
		pmid,
	)

	// #nosec G107
	resp, err := http.Get(efetchURL)
	if err != nil {
		return nil, fmt.Errorf("error making efetch request: %w", err)
	}
	defer resp.Body.Close()

	articleSet := &PubMedArticleSet{}
	if err := xml.NewDecoder(resp.Body).Decode(articleSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling efetch XML: %w", err)
	}

	return articleSet, nil
}

func fetchOADetails(pmcID string) (*OARecord, error) {
	oaURL := fmt.Sprintf(
		"https://www.ncbi.nlm.nih.gov/pmc/utils/oa/oa.fcgi?id=%s",
		pmcID,
	)

	// #nosec G107
	resp, err := http.Get(oaURL)
	if err != nil {
		return nil, fmt.Errorf("error making OA request: %w", err)
	}
	defer resp.Body.Close()

	oaResponse := &OAResponse{}
	if err := xml.NewDecoder(resp.Body).Decode(oaResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling OA response: %w", err)
	}
	if len(oaResponse.Records.Records) == 0 {
		return nil,
			fmt.Errorf("no records found in OA response for %s", pmcID)
	}
	// Assuming the first record is the one we want.
	return &oaResponse.Records.Records[0], nil
}

// PubMedArticleDownloader provides methods to download a PubMed article PDF.
type PubMedArticleDownloader struct {
	article *PubMedArticle
	pdfLink *OALink
}

// NewPubMedArticleDownloader creates a new downloader for a given PubMed ID.
// It fetches the article metadata upon creation.
func NewPubMedArticleDownloader(pmid string) (*PubMedArticleDownloader, error) {
	articleSet, err := fetchPubMedArticle(pmid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pubmed article %s: %w", pmid, err)
	}

	if len(articleSet.PubMedArticles) == 0 {
		return nil, fmt.Errorf("no PubMed article found for the given ID: %s", pmid)
	}

	return &PubMedArticleDownloader{
		article: &articleSet.PubMedArticles[0],
	}, nil
}

// IsPDFAvailable checks if a downloadable PDF is available for the article.
func (d *PubMedArticleDownloader) IsPDFAvailable() (bool, error) {
	if d.pdfLink != nil {
		return true, nil
	}

	pmcArticleID, found := Find(
		d.article.PubmedData.ArticleIdList.ArticleIds,
		isPMCID,
	)
	if !found {
		return false, nil
	}
	slog.Info("Found PMC ID", "pmcid", pmcArticleID.Value)

	oaRecord, err := fetchOADetails(pmcArticleID.Value)
	if err != nil {
		return false, fmt.Errorf("failed to fetch Open Access details: %w", err)
	}

	pdfLink, found := Find(oaRecord.Links, isPDFLink)
	if !found {
		return false, nil
	}

	d.pdfLink = pdfLink
	return true, nil
}

// DownloadPDF downloads the article's PDF to the specified file path.
// IsPDFAvailable should be called and return true before calling this method.
func (d *PubMedArticleDownloader) DownloadPDF(filePath string) error {
	if d.pdfLink == nil {
		return fmt.Errorf("PDF link not available or not checked. Call IsPDFAvailable first")
	}

	slog.Debug("Found PDF FTP link", "url", d.pdfLink.HREF)

	err := downloadFileFTP(d.pdfLink.HREF, filePath)
	if err != nil {
		return fmt.Errorf("failed to download file via FTP: %w", err)
	}
	slog.Info("Article download complete", "file", filePath)
	return nil
}

func downloadFileFTP(ftpURL string, filePath string) error {
	parsedURL, err := url.Parse(ftpURL)
	if err != nil {
		return fmt.Errorf("invalid FTP URL: %w", err)
	}

	host := parsedURL.Host
	if !strings.Contains(host, ":") {
		host += ":21" // Default FTP port
	}

	slog.Debug("Connecting to FTP server", "host", host)
	fclient, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}
	defer fclient.Quit()

	err = fclient.Login("anonymous", "anonymous")
	if err != nil {
		return fmt.Errorf("FTP login failed: %w", err)
	}
	slog.Debug("FTP login successful")

	path := parsedURL.Path
	slog.Info("Downloading file from FTP", "path", path, "savename", filePath)

	res, err := fclient.Retr(path)
	if err != nil {
		return fmt.Errorf("failed to retrieve file from FTP: %w", err)
	}
	defer res.Close()

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func formatAuthor(author Author) string {
	return fmt.Sprintf("%s %s", author.ForeName, author.LastName)
}

func isDOI(id ArticleID) bool {
	return id.IDType == "doi"
}

func isPMCID(id ArticleID) bool {
	return id.IDType == "pmc"
}

func isPDFLink(link OALink) bool {
	return link.Format == "pdf"
}
