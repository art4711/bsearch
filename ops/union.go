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
	h := un.peek()
	if h == nil {
		return nil
	}
	return h.CurrentDoc()
}

func (un *union) NextDoc(search *index.IbDoc) *index.IbDoc {
	d := un.CurrentDoc()
	// Chew up all documents bigger than search
	for d != nil && search.Less(d) {
		if un.peek().NextDoc(search) != nil {
			heap.Fix(un, 0)
		} else {
			heap.Pop(un)
		}
		d = un.CurrentDoc()
	}
	return d
}

// Len returns the number of elements in the union.
// Needed to implement heap.Interface.
func (un union) Len() int {
	return len(un)
}

// Compares two elements and returns which one is bigger than the other.
// Needed to implement heap.Interface.
// It's called Less because the heap is a min heap, but we want a max heap.
func (un union) Less(i, j int) bool {
	a := un[i].CurrentDoc()
	b := un[j].CurrentDoc()
	return b.Less(a) // we want a max-heap
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

// Peek at the head of the heap.
func (un union) peek() QueryOp {
	l := len(un)
	if l > 0 {
		return un[0]
	}
	return nil
}

func (un union) ProcessHeaders(hc HeaderCollector) {
	for _, n := range un {
		n.ProcessHeaders(hc)
	}
}
