package main

import (
	"bsearch/index"
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

	a1 := ops.NewAttr(in, "root:10")
	a2 := ops.NewAttr(in, "magic:boll")
	a3 := ops.NewAttr(in, "status:active")
	q := ops.NewIntersection(a1, a2, a3)

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
