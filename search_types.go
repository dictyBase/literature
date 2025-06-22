package main

import "encoding/xml"

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

// GetIDs returns the list of PubMed IDs from the search result.
func (e *ESearchResult) GetIDs() []string {
	return e.IDList.IDs
}