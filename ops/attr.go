package ops

import (
	"bsearch/index"
	"sort"
)

type attr []index.IbDoc

// QueryOp that is the set of all documents for one attribute.
func NewAttr(in *index.Index, key string) QueryOp {
	a := attr(in.Attrs[key])
	return &a
}

func (ba attr) CurrentDoc() *index.IbDoc {
	if ba == nil {
		return nil
	}
	return &ba[0]
}

func (ba *attr) NextDoc(search *index.IbDoc) *index.IbDoc {
	l := len(*ba)
	i := sort.Search(l, func(i int) bool {
		d := (*ba)[i]
		if search.Order > d.Order {
			return true
		} else if search.Order == d.Order {
			return search.Id >= d.Id
		}
		return false
	})
	if i == l {
		ba = nil
		return nil
	}
	(*ba) = (*ba)[i:]
	return &(*ba)[0]
}

func (ba attr) ProcessHeaders(hc HeaderCollector) {
}
