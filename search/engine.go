// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package main

import (
	"bsearch/index"
	"bsearch/parser"
	"flag"
	"fmt"
	"os"
	"log"
 	"runtime/pprof"
	"net"
	"time"
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

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
var listenport = flag.String("listen", "", "TCP port to listen to")
 
func main() {
	flag.Parse()
	if flag.NArg() != 1 || listenport == nil {
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

	ln, err := net.Listen("tcp", *listenport)
	if err != nil {
		log.Fatal("listen: %v\n", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("accept %v\n", err)
		}
		go handle(conn, in)
	}
}

func handle(conn net.Conn, in *index.Index) {
	b := [2048]byte{}
	n, err := conn.Read(b[:])
	if err != nil {
		log.Fatal("conn.Read %v\n", err)
	}
	bq := b[:n]
	log.Printf("query: %v\n", string(bq))
	t1 := time.Now()
	p := parser.Parse(in, string(bq))
	q := p.Stack[0]

	docarr := make([]*index.IbDoc, 0)

	s := index.NullDoc()
	for true {
		d := q.NextDoc(s)
		if d == nil {
			break
		}
		docarr = append(docarr, d)
		*s = *d
		s.Inc()
	}
	h := make(headers)
	q.ProcessHeaders(h)
	t2 := time.Now()

	log.Printf("Query time: %v\n", t2.Sub(t1))

	for k, v := range h {
		conn.Write([]byte(fmt.Sprintf("info:%v:%v\n", k, v)))
	}
	conn.Write([]byte(in.Header()))
	conn.Write([]byte("\n"))
	for o, d := range docarr {
		if d == nil {
			log.Fatalf("nil in docarr at %v", o)
		}
		doc, exists := in.Docs[d.Id]
		if !exists {
			log.Printf("Doc %v does not exist", d.Id)
		}
		conn.Write(doc)
		conn.Write([]byte("\n"))
	}
	conn.Close()
}
