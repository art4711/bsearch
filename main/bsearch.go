package main

import (
	"bsearch/index"
	"bsearch/parser"
	"bsearch/ops"
	"fmt"
	"os"
	"flag"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: bsearch <path to index blob>\n")
	flag.PrintDefaults()
	os.Exit(1)
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

	p := parser.Parse(in, "0 lim:10 root:10 OR magic:boll status:active")
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
}
