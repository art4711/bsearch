// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package parser

import (
	"bsearch/index"
	"bsearch/ops"
	"log"
	"strconv"
)

type Query struct {
	Stack []ops.QueryContainer
	in    *index.Index
}

func (q *Query) Limit(l string) {
	li, err := strconv.ParseInt(l, 10, 32)
	if err != nil {
		log.Fatal("limit parse failed")
	}
	q.push(ops.NewLimit(uint(li)))
}

func (q *Query) Offset(o string) {
	oi, err := strconv.ParseInt(o, 10, 32)
	if err != nil {
		log.Fatal("limit parse failed")
	}
	q.push(ops.NewOffset(uint(oi)))
}

func (q *Query) StartIntersection() {
	q.push(ops.NewIntersection())
}

func (q *Query) StartUnion() {
	q.push(ops.NewUnion())
}

func (q *Query) CountAll(s string) {
	q.push(ops.CountAll(s))
}

func (q *Query) PopAdd() {
	q.add(q.pop())
}

func (q *Query) push(o ops.QueryContainer) {
	q.Stack = append(q.Stack, o)
}

func (q *Query) pop() ops.QueryContainer {
	l := len(q.Stack) - 1
	r := q.Stack[l]
	q.Stack = q.Stack[:l]
	return r
}

func (q *Query) add(o ops.QueryOp) {
	q.Stack[len(q.Stack)-1].Add(o)
}

func (q *Query) Attr(a string) {
	q.add(ops.NewAttr(q.in, a))
}

func (q *Query) SetIndex(i *index.Index) {
	q.in = i
}

func Parse(i *index.Index, s string) Query {
	q := &QueryParser{Buffer: s}

	q.SetIndex(i)
	q.Init()

	if err := q.Parse(); err != nil {
		log.Fatal(err)
	}
	q.Execute()
	return q.Query
}
