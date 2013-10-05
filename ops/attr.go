package ops

import (
	"bsearch/index"
	"sort"
)

type attr struct {
	docs    []index.IbDoc
}

// QueryOp that is the set of all documents for one attribute.
func NewAttr(in *index.Index, key string) QueryOp {
	a := in.Attrs[key]
	if a == nil {
		return nil
	}
	return &attr{in.Attrs[key]}
}

func (ba *attr) CurrentDoc() *index.IbDoc {
	if ba.docs == nil {
		return nil
	}
	return &ba.docs[0]
}

func (ba *attr) NextDoc(search *index.IbDoc) *index.IbDoc {
	if search == nil {
		return ba.CurrentDoc()
	}

	l := len(ba.docs)
	i := sort.Search(l, func (i int) bool {
		d := ba.docs[i]
		if search.Order > d.Order {
			return true
		} else if search.Order == d.Order {
			return search.Id >= d.Id
		}
		return false
	});
	if i == l {
		ba.docs = nil
		return nil
	}
	ba.docs = ba.docs[i:]
	return &ba.docs[0]
}
