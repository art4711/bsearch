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
	"github.com/art4711/timers"
	"bufio"
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
	timer *timers.Timer
	
}

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
 
func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	s := engineState{ conf: make(bconf.Bconf) }
	s.timer = timers.New()

	timerConf := s.timer.Start("loadconf")
	s.conf.LoadConfFile(flag.Arg(0))
	timerConf.Stop()

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
	timerIndex := s.timer.Start("indexOpen")
	in, err := index.Open(dbname)
	timerIndex.Stop()
	if err != nil {
		log.Fatal(os.Stderr, "bindex.Open: %v\n", err)
	}
	defer in.Close()
	s.index = in

	cchan := make(chan string)

	go s.control(cchan)
	go s.listener()
	for {
		con := <- cchan
		if con == "stop" {
			break;
		}
	}
}

func (s engineState) listener() {
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
	qt := s.timer.Start("query")
	defer qt.Stop()

	b := [2048]byte{}
	n, err := conn.Read(b[:])
	if err != nil {
		log.Fatal("conn.Read %v\n", err)
	}
	bq := b[:n]

	pt := qt.Start("parse")
	p := parser.Parse(s.index, string(bq))
	q := p.Stack[0]

	docarr := make([]*index.IbDoc, 0)

	pt = pt.Handover("performQuery")
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
	pt = pt.Handover("ProcessHeaders")
	h := make(headers)
	q.ProcessHeaders(h)

	pt = pt.Handover("writeHeaders")
	for k, v := range h {
		conn.Write([]byte(fmt.Sprintf("info:%v:%v\n", k, v)))
	}
	conn.Write([]byte(s.index.Header()))
	conn.Write([]byte("\n"))
	pt = pt.Handover("writeDocs")
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
	pt.Stop()
}

func (s engineState) control(cchan chan string) {
	commandport := s.conf.GetString("port", "command")

	ln, err := net.Listen("tcp", ":" + commandport)
	if err != nil {
		log.Fatal("listen: %v\n", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("accept %v\n", err)
		}
		w := bufio.NewWriter(conn)
		s.timer.Foreach(func (name []string, tot, avg, max, min time.Duration, cnt int) {
			var n string
			for k, v := range name {
				if k > 0 {
					n += "." + v;
				} else {
					n = v
				}
			}
			fmt.Fprintf(w, "%v.cnt: %v\n", n, cnt)
			fmt.Fprintf(w, "%v.tot: %v\n", n, tot)
			fmt.Fprintf(w, "%v.min: %v\n", n, min)
			fmt.Fprintf(w, "%v.avg: %v\n", n, avg)
			fmt.Fprintf(w, "%v.max: %v\n", n, max)
		})
		w.Flush()
		conn.Close()
	}
}
