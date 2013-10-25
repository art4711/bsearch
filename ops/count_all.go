package ops

import (
	"bsearch/index"
	"fmt"
	"log"
)

type count_all struct {
	count uint
	next  QueryOp
	name  string
}

func CountAll(name string) QueryContainer {
	return &count_all{name: name}
}

func (ca *count_all) Add(n ...QueryOp) {
	if ca.next != nil || len(n) != 1 {
		log.Fatal("count_all.Add multiple")
	}
	ca.next = n[0]
}

func (ca count_all) CurrentDoc() *index.IbDoc {
	return ca.next.CurrentDoc()
}

func (ca *count_all) NextDoc(s *index.IbDoc) *index.IbDoc {
	d := ca.next.NextDoc(s)
	if d != nil {
		ca.count++
	}
	return d
}

func (ca count_all) ProcessHeaders(hc HeaderCollector) {
	hc.Add(ca.name, fmt.Sprint(ca.count))
	ca.next.ProcessHeaders(hc)
}
