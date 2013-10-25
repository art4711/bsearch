package ops

import (
	"bsearch/index"
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

/*
	i := sort.Search(l, func(i int) bool {
		d := (*ba)[i]
		return search.Order > d.Order || (search.Order == d.Order && search.Id >= d.Id)
	})
 The code below is an expanded version of this call. Inlining gives us a 30% speedup.
*/
	i, j := 0, l
	for i < j {
		h := i + (j - i)/2
		d := (*ba)[h]
		if (search.Order > d.Order || (search.Order == d.Order && search.Id >= d.Id)) {
			j = h
		} else {
			i = h + 1
		}
	}
	/* End of inline expanded sort.Search */

	if i == l {
		ba = nil
		return nil
	}
	(*ba) = (*ba)[i:]
	return &(*ba)[0]
}

func (ba attr) ProcessHeaders(hc HeaderCollector) {
}
