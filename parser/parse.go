// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package parser

import (
	"bsearch/index"
	"bsearch/ops"
	"bsearch/parser/opers"
	"bsearch/parser/classic"
	"bsearch/parser/structured"
	"github.com/art4711/timers"
)

func ParseClassic(s string) (*opers.Op, []error) {
	q := &classic.Parser{Buffer: s}

	q.Init()
	if err := q.Parse(); err != nil {
		return nil, append(q.Err, err)
	}
	q.Execute()

	if q.Err != nil {
		return nil, q.Err
	}

	return q.Stack[0], nil
}

func Classic(i *index.Index, s string) (ops.QueryOp, []error) {
	o, err := ParseClassic(s)
	if err != nil {
		return nil, err
	}
	return o.Generate(i)
}


func ParseStructured(s string, eet *timers.Event) (*opers.Op, []error) {
	q := &structured.Parser{Buffer: s}

	et := eet.Start("Init")
	q.Init()
	et = et.Handover("Parse")
	if err := q.Parse(); err != nil {
		return nil, append(q.Err, err)
	}
	et = et.Handover("Execute")
	defer et.Stop()
	q.Execute()

	if q.Err != nil {
		return nil, q.Err
	}

	return q.Stack[0], nil
}

func Structured(i *index.Index, s string, eet *timers.Event) (ops.QueryOp, []error) {
	et := eet.Start("Structured")
	o, err := ParseStructured(s, et)
	if err != nil {
		return nil, err
	}
	et = et.Handover("Generate")
	defer et.Stop()
	return o.Generate(i)
}

