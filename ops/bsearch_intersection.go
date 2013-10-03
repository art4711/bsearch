package ops

import (
	"bsearch/index"
)

type intersection struct {
	nodes []QueryOp
}

func NewIntersection(n ...QueryOp) QueryContainer {
	var it intersection

	it.Add(n...)

	return &it
}

func (it *intersection) Add(n ...QueryOp) {
	it.nodes = append(it.nodes, n...)
}

func (it *intersection) CurrentDoc() *index.IbDoc {
	if len(it.nodes) == 0 {
		return nil
	}
	return it.nodes[0].CurrentDoc()
}

func (it *intersection) NextDoc(search *index.IbDoc) *index.IbDoc {
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
