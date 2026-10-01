package main

import (
	"fmt"
	"io"

	IO "github.com/IBM/fp-go/v2/io"
)

var (
	// logWorkHeader prints the resolved work details before further steps.
	logWorkHeader = func(state State) IO.IO[State] {
		return func() State {
			out := writer(state)
			fmt.Fprintf(out, "Work: %s\n", state.Work.Title)
			printWorkFields(state)
			return state
		}
	}

	// logMetrics prints citation impact metrics.
	logMetrics = func(state State) IO.IO[State] {
		return func() State {
			out := writer(state)
			fmt.Fprintf(
				out,
				"Citations: %d | FWCI: %.2f | percentiles: %.2f-%.2f\n",
				state.Metrics.CitedByCount,
				state.Metrics.FWCI,
				state.Metrics.PercentileMin,
				state.Metrics.PercentileMax,
			)
			return state
		}
	}

	// logRefs prints the referenced-works summary.
	logRefs = func(state State) IO.IO[State] {
		return func() State {
			out := writer(state)
			fmt.Fprintf(out, "Cited works: %d\n", len(state.Refs))
			printRefList(state)
			return state
		}
	}

	// logCiting prints the citing-works page summary.
	logCiting = func(state State) IO.IO[State] {
		return func() State {
			out := writer(state)
			fmt.Fprintf(
				out,
				"Total citing works: %d | next cursor: %s\n",
				state.Citing.TotalCount,
				state.Citing.Cursor,
			)
			printCitingList(state)
			return state
		}
	}
)

// writer extracts the output writer from the state's logger.
func writer(st State) io.Writer {
	return st.Logger.Writer()
}

// printWorkFields prints identifiers, authors, and open access status.
func printWorkFields(st State) {
	work := st.Work
	out := writer(st)
	fmt.Fprintf(out, "  openalex: %s | doi: %s | year: %d\n",
		work.OpenAlexID, work.DOI, work.PublicationYear)
	fmt.Fprintf(out, "  cited_by: %d | type: %s | retracted: %t\n",
		work.CitedByCount, work.Type, work.IsRetracted)
	if len(work.Authors) > 0 {
		fmt.Fprintf(out, "  first author: %s (%s)\n",
			work.Authors[0].Author.DisplayName, work.Authors[0].Author.ORCID)
	}
	if work.OpenAccess.IsOA {
		fmt.Fprintf(out, "  open access: %s | %s\n",
			work.OpenAccess.OAStatus, work.OpenAccess.OAURL)
	}
}

// printRefList prints up to five referenced works.
func printRefList(st State) {
	out := writer(st)
	for i, ref := range st.Refs {
		if i >= 5 {
			fmt.Fprintf(out, "  ... and %d more\n", len(st.Refs)-i)
			break
		}
		fmt.Fprintf(out, "  - %s (%d)\n", ref.Title, ref.PublicationYear)
	}
}

// printCitingList prints the citing-works page entries.
func printCitingList(st State) {
	out := writer(st)
	for _, citing := range st.Citing.Results {
		fmt.Fprintf(out, "  - %s (%d) | cited_by: %d\n",
			citing.Title, citing.PublicationYear, citing.CitedByCount)
	}
}
