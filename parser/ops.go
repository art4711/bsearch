// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package parser

import (
	"strconv"
	"fmt"
	"errors"
)

type optype int
const (
	oAttr optype = iota
	oUnion
	oIntersection
	oOffset
	oLimit
	oCountAll
)

type op struct {
	typ optype

	// counter name, attribute name, etc.
	name string

	// only one of the following can be set. counter attrs, attr range values, etc.
	value []string
	intValue []int64
	opValue []*op

	contents []*op
}

type Query struct {
	stack []*op
	err error
}

var ErrSyntax = errors.New("query syntax error")
var ErrLimitRange = errors.New("limit out of range")
var ErrOffsetRange = errors.New("offset out of range")

func (q *Query) lim(l string) {
	li, err := strconv.ParseInt(l, 10, 32)
	if err != nil {
		q.err = ErrLimitRange
	}
	q.push(&op{ typ: oLimit, intValue: []int64{ li } })
}

func (q *Query) off(o string) {
	oi, err := strconv.ParseInt(o, 10, 32)
	if err != nil {
		q.err = ErrOffsetRange
	}
	q.push(&op{ typ: oOffset, intValue: []int64{ oi } })
}

func (q *Query) inter() {
	q.push(&op{ typ: oIntersection })
}

func (q *Query) union() {
	q.push(&op{ typ: oUnion })
}

func (q *Query) countall(s string) {
	q.push(&op{ typ: oCountAll, name: s})
}

func (q *Query) attr(a string) {
	q.add(&op{ typ: oAttr, name: a})		// split into name+value later.
}

func (q *Query) pa() {
	q.add(q.pop())
}

func (q *Query) push(o *op) {
	q.stack = append(q.stack, o)
}

func (q *Query) pop() *op {
	l := len(q.stack) - 1
	r := q.stack[l]
	q.stack = q.stack[:l]
	return r
}

func (q *Query) add(o *op) {
	top := q.stack[len(q.stack)-1]
	top.contents = append(top.contents, o)
}

func (o op) String() string {
	var t string
	switch o.typ {
	case oAttr:	t = "attr"
	case oUnion:	t = "union"
	case oIntersection:	t = "intersection"
	case oOffset:	t = "offset"
	case oLimit:	t = "limit"
	case oCountAll:	t = "count_all"
	}
	s := "(" + t
	if o.name != "" {
		s += ` "` + o.name + `"`
	}
	var vs string
	for _, v := range o.value {
		vs += ` "` + v + `"`
	}
	for _, v := range o.intValue {
		vs += ` ` + fmt.Sprint(v)
	}
	for _, v := range o.opValue {
		vs += v.String()
	}
	if vs != "" {
		s += ` [` + vs + ` ]`
	}
	for _, v := range o.contents {
		s += " " + v.String()
	}
	s += ")"
	return s
}
