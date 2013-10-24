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

func stdtest(in *index.Index) {
	p := parser.Parse(in, "0 lim:10 count_all(hej) root:10 OR magic:boll status:active")
	q := p.Stack[0]

	ops.Dump(q, 0)

	fmt.Printf("%v\n", in.Header())

	d := index.NullDoc()
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

func bltest(in *index.Index) {
	p := parser.Parse(in, "0 count_all(hej) region:11 category:1000 OR category:2000 OR category:3000")
	q := p.Stack[0]

	ops.Dump(q, 0)

	d := index.NullDoc()
	for true {
		d = q.NextDoc(d)
		if d == nil {
			break
		}
		d = d.Inc()
	}
	h := make(headers)
	q.ProcessHeaders(h)
	for k, v := range h {
		fmt.Printf("info:%v:%v\n", k, v)
	}	
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

	bltest(in)
}
