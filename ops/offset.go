package ops

import (
	"bsearch/index"
	"log"
)

type offset struct {
	offset uint
	next   QueryOp
}

func NewOffset(o uint) QueryContainer {
	return &offset{offset: o}
}

func (o *offset) Add(n ...QueryOp) {
	if o.next != nil || len(n) != 1 {
		log.Fatal("offset.Add multiple")
	}
	o.next = n[0]
}

func (o offset) CurrentDoc() *index.IbDoc {
	d := o.next.CurrentDoc()
	if o.offset == 0 {
		return d
	}
	var s index.IbDoc
	s = *d
	for ; o.offset > 0; o.offset-- {
		s.Inc()
		d := o.next.NextDoc(&s)
		s = *d
	}
	return o.next.CurrentDoc()

}

func (o *offset) NextDoc(s *index.IbDoc) *index.IbDoc {
	for ; o.offset > 0; o.offset-- {
		d := o.next.NextDoc(s)
		*s = *d
		s.Inc()
	}
	return o.next.NextDoc(s)
}

func (o offset) ProcessHeaders(hc HeaderCollector) {
	o.next.ProcessHeaders(hc)
}
