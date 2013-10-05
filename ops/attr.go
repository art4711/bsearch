package ops

import (
	"bsearch/index"
	"sort"
)

type attr struct {
	docs    []index.IbDoc
	currptr int
}

// QueryOp that is the set of all documents for one attribute.
func NewAttr(in *index.Index, key string) QueryOp {
	a := in.Attrs[key]
	if a == nil {
		return nil
	}
	return &attr{in.Attrs[key], 0}
}

func (ba *attr) CurrentDoc() *index.IbDoc {
	if ba.currptr == -1 {
		return nil
	}
	return &ba.docs[ba.currptr]
}

func (ba *attr) NextDoc(search *index.IbDoc) *index.IbDoc {
	if search == nil {
		r := &ba.docs[ba.currptr]
		ba.currptr++
		return r
	}

	from := ba.currptr
	l := len(ba.docs) - from
	i := sort.Search(l, func (i int) bool {
		d := ba.docs[from + i]
		if search.Order > d.Order {
			return true
		} else if search.Order == d.Order {
			return search.Id >= d.Id
		}
		return false
	});
	if i == l {
		ba.currptr = -1
		return nil
	}
	ba.currptr = from + i
	return &ba.docs[from + i]
}
