package ops

import (
	"bsearch/index"
)

type HeaderCollector interface {
	Add(key, value string)
}

// QueryOp is the interface for search queries implemented by everything
// that will return documents. Each QueryOp is assumed to define access
// to a sorted set of documents sorted on IbDoc.order.
type QueryOp interface {

	// CurrentDoc returns the last document returned by NextDoc
	// or the first document from this query if NextDoc hasn't been
	// called yet.
	CurrentDoc() *index.IbDoc

	// NextDoc returns the document equal to `search` or next higher.
	NextDoc(search *index.IbDoc) *index.IbDoc

	// Recursively adds any headers this might need to return.
	ProcessHeaders(hc HeaderCollector)
}

// QueryContainer is an interface for ops that not only implement sets of
// documents like QueryOp, but are also containers for other queries.
// This applies to intersections and unions.
type QueryContainer interface {
	QueryOp
	// Add adds one or more QueryOp to the container.
	Add(...QueryOp)
}
