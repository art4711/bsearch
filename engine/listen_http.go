// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package engine

import (
	"bsearch/parser"
	"fmt"
	"net/http"
	"encoding/json"
	"log"
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

	result := make(map[string]interface{})
	resultInfo := make(headers)

	et.Handover("parse")
	q, err := parser.Structured(s.Index, req.FormValue("q"))
	if err != nil {
		resultInfo.Add("error", "parse error")
		resultInfo.Add("parse_error", fmt.Sprint(err))
	} else {
		et.Handover("perform")
		docarr := performQuery(q, et)

		et = et.Handover("ProcessHeaders")
		q.ProcessHeaders(resultInfo)

		docsData := make(map[string]interface{})
		for _, d := range docarr {
			docsData[fmt.Sprint(d.Id)] = s.Index.SplitDoc(d.Id)
		}
		result["docs"] = docsData
	}

	result["info"] = resultInfo
	json, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Printf("HandleHTTPQuery: json.Marshal: %v", err)
	}
	w.Write(json)
}
