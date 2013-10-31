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
	"github.com/art4711/bconf"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: bsearch <path to config>\n")
	flag.PrintDefaults()
	os.Exit(1)
}

type headers map[string]string

func (h headers) Add(k, v string) {
	h[k] = v
}

type engineState struct {
	conf bconf.Bconf
	index *index.Index
}

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
 
func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	s := engineState{}

	s.conf = make(bconf.Bconf)
	s.conf.LoadConfFile(flag.Arg(0))

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	dbname := s.conf.GetString("db_name")
	in, err := index.Open(dbname)
	if err != nil {
		log.Fatal(os.Stderr, "bindex.Open: %v\n", err)
	}
	defer in.Close()
	s.index = in

	listenport := s.conf.GetString("port", "search")

	ln, err := net.Listen("tcp", ":" + listenport)
	if err != nil {
		log.Fatal("listen: %v\n", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("accept %v\n", err)
		}
		go s.handle(conn)
	}
}

func (s engineState) handle(conn net.Conn) {
	b := [2048]byte{}
	n, err := conn.Read(b[:])
	if err != nil {
		log.Fatal("conn.Read %v\n", err)
	}
	bq := b[:n]
	log.Printf("query: %v\n", string(bq))
	t1 := time.Now()
	p := parser.Parse(s.index, string(bq))
	q := p.Stack[0]

	docarr := make([]*index.IbDoc, 0)

	search := index.NullDoc()
	for true {
		d := q.NextDoc(search)
		if d == nil {
			break
		}
		docarr = append(docarr, d)
		*search = *d
		search.Inc()
	}
	h := make(headers)
	q.ProcessHeaders(h)
	t2 := time.Now()

	log.Printf("Query time: %v\n", t2.Sub(t1))

	for k, v := range h {
		conn.Write([]byte(fmt.Sprintf("info:%v:%v\n", k, v)))
	}
	conn.Write([]byte(s.index.Header()))
	conn.Write([]byte("\n"))
	for o, d := range docarr {
		if d == nil {
			log.Fatalf("nil in docarr at %v", o)
		}
		doc, exists := s.index.Docs[d.Id]
		if !exists {
			log.Printf("Doc %v does not exist", d.Id)
		}
		conn.Write(doc)
		conn.Write([]byte("\n"))
	}
	conn.Close()
}
