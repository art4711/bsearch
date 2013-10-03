package ops

import (
	"bsearch/index"
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
	from := ba.currptr
	to := len(ba.docs) - 1

	if search == nil {
		r := &ba.docs[ba.currptr]
		ba.currptr++
		return r
	}

	for true {
		middle := from + (to-from)/2
		d := &ba.docs[middle]
		c := search.Cmp(d)
		if c == 0 {
			ba.currptr = middle
			return d
		}
		if to == from {
			if c > 0 {
				ba.currptr = to
				return d
			}
			break
		}
		if c < 0 {
			from = middle + 1
		} else {
			if from == middle {
				ba.currptr = middle
				return d
			}
			to = middle
		}
	}
	ba.currptr = -1
	return nil
}
