package main

import (
	"bsearch/index"
	"bsearch/ops"
	"bsearch/parser"
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: bsearch <path to index blob>\n")
	flag.PrintDefaults()
	os.Exit(1)
}

type headers map[string]string

func (h headers) Add(k, v string) {
	h[k] = v
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}
	dbname := flag.Arg(0)
	in, err := index.Open(dbname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bindex.Open: %v\n", err)
		return
	}
	defer in.Close()

	p := parser.Parse(in, "0 lim:10 count_all(hej) root:10 OR magic:boll status:active")
	q := p.Stack[0]

	ops.Dump(q, 0)

	fmt.Printf("%v\n", in.Header())

	var d *index.IbDoc
	for true {
		d = q.NextDoc(d)
		if d == nil {
			break
		}
		fmt.Printf("%v\n", string(in.Docs[d.Id]))
		d = d.Inc()
	}
	h := make(headers)
	q.ProcessHeaders(h)
	for k, v := range h {
		fmt.Printf("info:%v:%v\n", k, v)
	} 
}
