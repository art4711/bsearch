package ops

import (
	"bsearch/index"
//	"sort"
)

type attr []index.IbDoc

// QueryOp that is the set of all documents for one attribute.
func NewAttr(in *index.Index, key string) QueryOp {
	a := attr(in.Attrs[key])
	return &a
}

func (ba attr) CurrentDoc() *index.IbDoc {
	return &ba[0]
}

func (ba *attr) NextDoc(search *index.IbDoc) *index.IbDoc {
	/*
	 * We know that quite often the doc we're looking for is quite often early in the attribute.
	 * We abuse that knowledge to make a linear scan of the first few elements of the doc array
	 * to see if we can catch it.
	 *
	 * The constant of 5 was determined experimentally to be good enough without risking too much
	 * in the pessimal case (it's also related to a cache line size). In an optimal world we should
	 * probably try to determine it run-time.
	 *
	 * We also bias the binary search to the left, potentially degenerating into a linear search on
	 * small attributes, but the experimental number 16 gives the best performance on test data.
	 */

	const firstLinear = 5
	const leftBias = 16

	l := len(*ba)

	start := 0
	for start = 0; start < l && start < firstLinear; start++ {
		if (*ba)[start].LessEqual(*search) {
			(*ba) = (*ba)[start:]
			return &(*ba)[0]
		}
	}
	
/*
	i := sort.Search(l, func(i int) bool {
		d := (*ba)[i]
		return d.LessEqual(*search)
	})
 The code below is an expanded version of this call. Inlining gives us a large speedup and allows us to cheat.
*/
	i, j := start, l
	for i < j {
		h := i + (j - i)/leftBias
		d := (*ba)[h]
		if d.LessEqual(*search) {
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
