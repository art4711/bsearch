// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package parser

import (
	"bsearch/index"
	"bsearch/ops"
)

func ParseClassic(s string) (*op, error) {
	q := &ClassicParser{Buffer: s}

	q.Init()
	if err := q.Parse(); err != nil {
		return nil, ErrSyntax
	}
	q.Execute()

	if q.err != nil {
		return nil, q.err
	}

	return q.stack[0], nil
}

func Classic(i *index.Index, s string) (ops.QueryOp, error) {
	o, err := ParseClassic(s)
	if err != nil {
		return nil, err
	}
	return o.Generate(i)
}

/*
func ParseStructured(s string) (*op, error) {
	q := &structured.StructuredParser{Buffer: s}

	q.Init()
	if err := q.Parse(); err != nil {
		return nil, ErrSyntax
	}
	q.Execute()

	if q.err != nil {
		return nil, q.err
	}

	return q.stack[0], nil
}

func Structured(i *index.Index, s string) (ops.QueryOp, error) {
	o, err := ParseClassic(s)
	if err != nil {
		return nil, err
	}
	return o.generate(i)
}
*/
