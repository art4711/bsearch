package engine

import (
	"net/http"
	"time"
	"encoding/json"
	"fmt"
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

	s.Timer.Foreach(func (name []string, tot, avg, max, min time.Duration, cnt int) {
		dp := data
		for _, v := range name {
			if _, exists := dp[v]; !exists {
				dp[v] = make(map[string]interface{})
			}
			dp = dp[v].(map[string]interface{})
		}
		dp["cnt"] = fmt.Sprint(cnt)
		dp["tot"] = tot.String()
		dp["min"] = tot.String()
		dp["avg"] = tot.String()
		dp["max"] = tot.String()
	})
	json, _ := json.MarshalIndent(data, "", "    ")
	w.Write(json)
}