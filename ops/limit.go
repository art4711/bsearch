package ops

import (
	"bsearch/index"
	"log"
)

type limit struct {
	lim  uint
	next QueryOp
}

func NewLimit(lim uint) QueryContainer {
	return &limit{lim: lim}
}

func (l *limit) Add(n ...QueryOp) {
	if l.next != nil || len(n) != 1 {
		log.Fatal("limit.Add multiple")
	}
	l.next = n[0]
}

func (l *limit) CurrentDoc() *index.IbDoc {
	if l.lim == 0 {
		return nil
	}
	return l.next.CurrentDoc()
}

func (l *limit) NextDoc(s *index.IbDoc) *index.IbDoc {
	if l.lim == 0 {
		return nil
	}
	l.lim--
	return l.next.NextDoc(s)
}


func (l limit) ProcessHeaders(hc HeaderCollector) {
	l.next.ProcessHeaders(hc)
}
