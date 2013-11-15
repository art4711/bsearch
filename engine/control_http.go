// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package engine

import (
	"net/http"
	"github.com/art4711/timers"
)

func (s EngineState) ControlHTTP(cchan chan string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stop", func(w http.ResponseWriter, req *http.Request) {
		cchan <- "stop"
	})
	mux.HandleFunc("/timers", func (w http.ResponseWriter, req *http.Request) {
		s.Timer.JSONHandler(w, req)
	})
	mux.HandleFunc("/graph", func (w http.ResponseWriter, req *http.Request) {
		timers.JSONHandlerGraph(w, req, "/timers")
	})

	addr := ":" + s.Conf.GetString("port", "command")
	hs := &http.Server{
		Addr: addr,
		Handler: mux,
	}
	hs.ListenAndServe()
}

