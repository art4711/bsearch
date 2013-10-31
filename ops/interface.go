// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
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

	// CurrentDoc returns the last document returned by NextDoc.
	// Can't be called on invalid or exhausted QueryOp or a QueryOp
	// that hasn't been used with NextDoc yet.
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
