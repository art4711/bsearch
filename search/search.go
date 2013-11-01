// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package main

import (
	"bsearch/index"
	"bsearch/engine"
	"flag"
	"fmt"
	"os"
	"log"
 	"runtime/pprof"
	"github.com/art4711/bconf"
	"github.com/art4711/timers"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: bsearch <path to config>\n")
	flag.PrintDefaults()
	os.Exit(1)
}

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
 
func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	s := engine.EngineState{ Conf: make(bconf.Bconf) }
	s.Timer = timers.New()

	timerConf := s.Timer.Start("loadconf")
	s.Conf.LoadConfFile(flag.Arg(0))
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

	dbname := s.Conf.GetString("db_name")
	timerIndex := s.Timer.Start("indexOpen")
	in, err := index.Open(dbname)
	timerIndex.Stop()
	if err != nil {
		log.Fatal(os.Stderr, "bindex.Open: %v\n", err)
	}
	defer in.Close()
	s.Index = in

	cchan := make(chan string)

	go s.ControlHTTP(cchan)
	go s.Listener()
	for {
		con := <- cchan
		if con == "stop" {
			break;
		}
	}
}
