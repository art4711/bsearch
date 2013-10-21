package main

import (
	"bsearch/index"
//	"bsearch/ops"
	"bsearch/parser"
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

	p := parser.Parse(in, "17 lim:42 root:10 OR magic:boll")
/*	fmt.Printf("%v\n", p)

	a1 := ops.NewAttr(in, "root:10")
	a2 := ops.NewAttr(in, "magic:boll")
	a3 := ops.NewAttr(in, "status:active")
	q := ops.NewIntersection(ops.NewUnion(a1, a2), a3)
	q.Add(a3)*/
	q := p.Stack[0]

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
