// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package engine

import (
	"bsearch/parser"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
)

func (s EngineState) ListenHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/x", s.HandleHTTPQuery)

	addr := ":" + s.Conf.GetString("port", "http_search")
	hs := http.Server{
		Addr: addr,
		Handler: mux,
	}
	hs.ListenAndServe()

}

func (s EngineState) HandleHTTPQuery(w http.ResponseWriter, req *http.Request) {
	qt := s.Timer.Start("httpquery")
	defer qt.Stop()

	et := qt.Start("req")
	qv := req.FormValue("q")
	et.Handover("parse")
	q, err := parser.Structured(s.Index, qv)
	if err != nil {
		log.Printf("parser error [%v]: %v", qv, err)
		w.Write([]byte("parser failed:\n"))
		w.Write([]byte(fmt.Sprint(err)))
		w.Write([]byte("\n"))
		et.Stop()
		return
	}

	et.Handover("perform")
	docarr := performQuery(q, et)

	et = et.Handover("ProcessHeaders")
	h := make(headers)
	q.ProcessHeaders(h)

	docsData := make(map[string]interface{})
	for _, d := range docarr {
		docsData[fmt.Sprint(d.Id)] = s.Index.SplitDoc(d.Id)
	}

	data := make(map[string]interface{})
	data["info"] = h
	data["docs"] = docsData	

	json, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("json err: %v", err)
	}
	w.Write(json)
}
