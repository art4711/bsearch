// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package main

import (
	"bsearch/index"
	"bsearch/ops"
	"bsearch/parser"
	"flag"
	"fmt"
	"os"
	"time"
	"log"
 	"runtime/pprof"
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

	s := index.NullDoc()
	for true {
		d := q.NextDoc(s)
		if d == nil {
			break
		}
		fmt.Printf("%v\n", string(in.Docs[d.Id]))
		*s = *d
		s.Inc()
	}
	h := make(headers)
	q.ProcessHeaders(h)
	for k, v := range h {
		fmt.Printf("info:%v:%v\n", k, v)
	}
}

func bltest(in *index.Index) {
	p := parser.Parse(in, "0 count_all(hej) status:active region:11 category:1000 OR category:2000")
	q := p.Stack[0]

//	t1 := time.Now()
	s := index.NullDoc()
	for true {
		d := q.NextDoc(s)
		if d == nil {
			break
		}
		*s = *d
		s.Inc()
	}
	h := make(headers)
	q.ProcessHeaders(h)
//	t2 := time.Now()
//	fmt.Printf("rt: %v\n", t2.Sub(t1))
}

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
 
func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}
	dbname := flag.Arg(0)
	in, err := index.Open(dbname)
	if err != nil {
		log.Fatal(os.Stderr, "bindex.Open: %v\n", err)
	}
	defer in.Close()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	bltest(in)	// Warm up.
	t1 := time.Now()
	for i := 0; i < 10; i++ {
		bltest(in)
	}
	t2 := time.Now()
	fmt.Printf("t: %v\n", t2.Sub(t1))
}
