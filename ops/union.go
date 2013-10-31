// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package ops

import (
	"bsearch/index"
	"container/heap"
)

type union []QueryOp

func NewUnion(n ...QueryOp) QueryContainer {
	var un union

	heap.Init(&un)

	un.Add(n...)
	return &un
}

func (un *union) Add(nodes ...QueryOp) {
	for _, n := range nodes {
		if n.CurrentDoc() != nil {
			heap.Push(un, n)
		}
	}
}

func (un union) CurrentDoc() *index.IbDoc {
	return un[0].CurrentDoc()
}

func (un *union) NextDoc(search *index.IbDoc) *index.IbDoc {
	d := (*un)[0].CurrentDoc()
	// Chew up all documents bigger than search
	for d != nil && search.Less(*d) {
		if (*un)[0].NextDoc(search) != nil {
			// We modified the head element in the heap, fix it.
			heap.Fix(un, 0)
		} else {
			// The head element was emptied, remove it.
			heap.Pop(un)
		}
		d = un.CurrentDoc()
	}
	return d
}
/*
// This is slower than version above, but doesn't use CurrentDoc (which is problematic).
func (un *union) NextDoc(search *index.IbDoc) *index.IbDoc {
	for len(*un) > 0 {
		first := (*un)[0]
		d := first.NextDoc(search)
		if d != nil {
			heap.Fix(un, 0)
			if (*un)[0] == first {
				return d
			}
		} else {
			heap.Pop(un)
		}
	}
	return nil
}
*/

// Len returns the number of elements in the union.
// Needed to implement heap.Interface.
func (un union) Len() int {
	return len(un)
}

// Compares two elements and returns which one is bigger than the other.
// Needed to implement heap.Interface.
func (un union) Less(i, j int) bool {
	a := un[i].CurrentDoc()
	b := un[j].CurrentDoc()
	return b.Less(*a) // we want a max-heap
}

// Swap to elements in the union.
// Needed to implement heap.Interface.
func (un union) Swap(i, j int) {
	un[i], un[j] = un[j], un[i]
}

// Add an element at the end of the union.
// Needed to implement heap.Interface.
func (un *union) Push(x interface{}) {
	*un = append(*un, x.(QueryOp))
}

// Remove one element from the end of the union.
// Needed to implement heap.Interface.
func (un *union) Pop() interface{} {
	l := len(*un)
	r := (*un)[l-1]
	*un = (*un)[:l-1]
	return r
}

func (un union) ProcessHeaders(hc HeaderCollector) {
	for _, n := range un {
		n.ProcessHeaders(hc)
	}
}
