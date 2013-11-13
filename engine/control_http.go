// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package engine

import (
	"net/http"
	"encoding/json"
	"fmt"
	"github.com/art4711/timers"
)

func (s EngineState) ControlHTTP(cchan chan string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stop", func(w http.ResponseWriter, req *http.Request) {
		cchan <- "stop"
	})
	mux.HandleFunc("/timers", s.handleTimers)
	addr := ":" + s.Conf.GetString("port", "command")
	hs := &http.Server{
		Addr: addr,
		Handler: mux,
	}
	hs.ListenAndServe()
}

func (s EngineState) handleTimers(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	req.ParseForm()
	var filt string
	if len(req.Form["filt"]) > 0 && len(req.Form["filt"][0]) > 0 {
		filt = req.Form["filt"][0]
	}

	s.Timer.Foreach(func (name []string, cnt *timers.Counts) {
		if filt != "" && len(name) > 0 && name[0] != filt {
			return
		}
		dp := data
		for _, v := range name {
			if _, exists := dp[v]; !exists {
				dp[v] = make(map[string]interface{})
			}
			dp = dp[v].(map[string]interface{})
		}
		dp["cnt"] = fmt.Sprint(cnt.Count)
		dp["tot"] = cnt.Tot.String()
		dp["min"] = cnt.Min.String()
		dp["avg"] = cnt.Avg.String()
		dp["max"] = cnt.Max.String()
		dp["numgc"] = fmt.Sprint(cnt.NumGC)
		dp["bytes"] = fmt.Sprint(cnt.BytesAlloc)
		dp["gctime"] = fmt.Sprint(cnt.GCTime)
	})
	json, _ := json.MarshalIndent(data, "", "    ")
	w.Write(json)
}