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

// QueryContainer
func (un *union) Add(nodes ...QueryOp) {
	for _, n := range(nodes) {
		heap.Push(un, n)
	}
}

// QueryOp
func (un *union) CurrentDoc() *index.IbDoc {
	h := un.head()
	if h == nil {
		return nil
	}
	return h.CurrentDoc()
}

// QueryOp
func (un *union) NextDoc(search *index.IbDoc) *index.IbDoc {
	d := un.CurrentDoc()
	if search == nil {
		return d
	}
	// Chew up all documents bigger than search
	for d != nil && search.Cmp(d) < 0 {
		n := heap.Pop(un).(QueryOp)
		// Only put the element back into the heap if it's not empty.
		if n.NextDoc(search) != nil {
			heap.Push(un, n)
		}
		d = un.CurrentDoc()
	}
	return d
}

// sort.Interface
func (un union) Len() int {
	return len(un)
}


// sort.Interface
func (un union) Less(i, j int) bool {
	a := un[i].CurrentDoc()
	b := un[j].CurrentDoc()
	return a.Cmp(b) > 0		// we want a max-heap
}

// sort.Interface
func (un union) Swap(i, j int) {
	un[i], un[j] = un[j], un[i]
}

// heap.Interface
func (un *union) Push(x interface{}) {
	*un = append(*un, x.(QueryOp))
}

// heap.Interface
func (un *union) Pop() interface{} {
	l := len(*un)
	r := (*un)[l - 1]
	*un = (*un)[:l-1]
	return r
}

// Returns the head of the heap.
func (un union) head() QueryOp {
	l := len(un)
	if l > 0 {
		return un[l - 1]
	}
	return nil
	
}