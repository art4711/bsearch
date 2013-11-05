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

	et := qt.Start("read")
	b := [2048]byte{}
	n, err := conn.Read(b[:])
	if err != nil {
		log.Fatal("conn.Read %v\n", err)
	}
	bq := b[:n]

	writer := bufio.NewWriter(conn)
	defer writer.Flush()

	et = et.Handover("parse")
	q, err := parser.Classic(s.Index, string(bq))
	if err != nil {
		fmt.Fprintf(writer, "info:error:%v\n", err)
		et.Stop()
		return
	}

	docarr := make([]*index.IbDoc, 0)

	et = et.Handover("performQuery")
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
	et = et.Handover("ProcessHeaders")
	h := make(headers)
	q.ProcessHeaders(h)

	et = et.Handover("writeHeaders")
	for k, v := range h {
		fmt.Fprintf(writer, "info:%v:%v\n", k, v)
	}
	writer.WriteString(s.Index.Header())
	et = et.Handover("writeDocs")
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
	et.Stop()
}
