package literature

// IsDOI checks if an ArticleID is a DOI.
func IsDOI(id ArticleID) bool {
	return id.IDType == "doi"
}

// IsPMCID checks if an ArticleID is a PMCID.
func IsPMCID(id ArticleID) bool {
	return id.IDType == "pmc"
}

// IsPDFLink checks if an OALink is a PDF download link.
func IsPDFLink(link OALink) bool {
	return link.Format == "pdf"
}
