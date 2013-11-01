// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package engine

import (
	"bsearch/index"
	"bsearch/parser"
	"fmt"
	"log"
	"net"
	"time"
	"github.com/art4711/bconf"
	"github.com/art4711/timers"
	"bufio"
)

type headers map[string]string

func (h headers) Add(k, v string) {
	h[k] = v
}

type EngineState struct {
	Conf bconf.Bconf
	Index *index.Index
	Timer *timers.Timer
	
}

func (s EngineState) Listener() {
	listenport := s.Conf.GetString("port", "search")

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

func (s EngineState) handle(conn net.Conn) {
	defer conn.Close()
	qt := s.Timer.Start("query")
	defer qt.Stop()

	b := [2048]byte{}
	n, err := conn.Read(b[:])
	if err != nil {
		log.Fatal("conn.Read %v\n", err)
	}
	bq := b[:n]

	writer := bufio.NewWriter(conn)
	defer writer.Flush()

	pt := qt.Start("parse")
	p := parser.Parse(s.Index, string(bq))
	q := p.Stack[0]

	docarr := make([]*index.IbDoc, 0)

	pt = pt.Handover("performQuery")
	search := index.NullDoc()
	for {
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
		fmt.Fprintf(writer, "info:%v:%v\n", k, v)
	}
	writer.WriteString(s.Index.Header())
	pt = pt.Handover("writeDocs")
	for o, d := range docarr {
		if d == nil {
			log.Fatalf("nil in docarr at %v", o)
		}
		doc, exists := s.Index.Docs[d.Id]
		if !exists {
			log.Printf("Doc %v does not exist", d.Id)
		}
		writer.Write(doc)
		writer.WriteString("\n")
	}
	pt.Stop()
}

func (s EngineState) Control(cchan chan string) {
	commandport := s.Conf.GetString("port", "command")

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
		s.Timer.Foreach(func (name []string, tot, avg, max, min time.Duration, cnt int) {
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
