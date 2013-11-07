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
)

func ParseClassic(s string) (*opers.Op, error) {
	q := &classic.Parser{Buffer: s}

	q.Init()
	if err := q.Parse(); err != nil {
		return nil, opers.ErrSyntax
	}
	q.Execute()

	if q.Err != nil {
		return nil, q.Err
	}

	return q.Stack[0], nil
}

func Classic(i *index.Index, s string) (ops.QueryOp, error) {
	o, err := ParseClassic(s)
	if err != nil {
		return nil, err
	}
	return o.Generate(i)
}


func ParseStructured(s string) (*opers.Op, error) {
	q := &structured.StructuredParser{Buffer: s}

	q.Init()
	if err := q.Parse(); err != nil {
		return nil, err
	}
	q.Execute()

	if q.Err != nil {
		return nil, q.Err
	}

	return q.Stack[0], nil
}

func Structured(i *index.Index, s string) (ops.QueryOp, error) {
	o, err := ParseStructured(s)
	if err != nil {
		return nil, err
	}
	return o.Generate(i)
}

