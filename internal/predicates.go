package internal

const (
	doiIDType = "doi"
	pmcIDType = "pmc"
)

// IsDOI checks if an ArticleID is a DOI.
func IsDOI(id ArticleID) bool {
	return id.IDType == doiIDType
}

// IsPMCID checks if an ArticleID is a PMCID.
func IsPMCID(id ArticleID) bool {
	return id.IDType == pmcIDType
}
