package ops

import (
	"bsearch/index"
)

type intersection []QueryOp

// QueryOp that is the intersection of the sets added to this container.
// QueryOps can be added to the container with this constructor or later with Add.
func NewIntersection(n ...QueryOp) QueryContainer {
	var it intersection

	it.Add(n...)

	return &it
}

func (it *intersection) Add(n ...QueryOp) {
	*it = append(*it, n...)
}

func (it intersection) CurrentDoc() *index.IbDoc {
	if len(it) == 0 {
		return nil
	}
	return (it)[0].CurrentDoc()
}

func (it intersection) NextDoc(search *index.IbDoc) *index.IbDoc {
	var d *index.IbDoc

	start_node := -1

	for true {
		for i, n := range it {
			if i == start_node {
				return d
			}
			d = n.NextDoc(search)
			if d == nil {
				return nil
			}
			if !search.Equal(d) {
				search = d
				start_node = i
			}
		}
		if start_node == -1 {
			return d
		}
	}
	return nil
}

func (it intersection) ProcessHeaders(hc HeaderCollector) {
	for _, n := range it {
		n.ProcessHeaders(hc)
	}
}
