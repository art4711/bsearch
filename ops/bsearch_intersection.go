package ops

import (
	"bsearch/index"
)

type Intersection struct {
	nodes []QueryOp
}

func NewIntersection(n ...QueryOp) *Intersection {
	var it Intersection

	it.Add(n...)

	return &it
}

func (it *Intersection) Add(n ...QueryOp) {
	it.nodes = append(it.nodes, n...)
}

func (it *Intersection) CurrentDoc() *index.IbDoc {
	if len(it.nodes) == 0 {
		return nil
	}
	return it.nodes[0].CurrentDoc()
}

func (it *Intersection) NextDoc(search *index.IbDoc) *index.IbDoc {
	var d *index.IbDoc

	start_node := -1

	for true {
		for i, n := range it.nodes {
			if i == start_node {
				return d
			}
			d = n.NextDoc(search)
			if d == nil {
				return nil
			}
			if search == nil || search.Cmp(d) != 0 {
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
