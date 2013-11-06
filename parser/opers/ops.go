// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package opers

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

type Op struct {
	typ optype

	// counter name, attribute name, etc.
	name string

	// only one of the following can be set. counter attrs, attr range values, etc.
	value []string
	intValue []int64
	opValue []*Op

	contents []*Op
}

type Query struct {
	Stack []*Op
	Err error
}

var ErrSyntax = errors.New("query syntax error")
var ErrLimitRange = errors.New("limit out of range")
var ErrOffsetRange = errors.New("offset out of range")

func (q *Query) Lim(l string) {
	li, err := strconv.ParseInt(l, 10, 32)
	if err != nil {
		q.Err = ErrLimitRange
	}
	q.push(&Op{ typ: oLimit, intValue: []int64{ li } })
}

func (q *Query) Off(o string) {
	oi, err := strconv.ParseInt(o, 10, 32)
	if err != nil {
		q.Err = ErrOffsetRange
	}
	q.push(&Op{ typ: oOffset, intValue: []int64{ oi } })
}

func (q *Query) Inter() {
	q.push(&Op{ typ: oIntersection })
}

func (q *Query) Union() {
	q.push(&Op{ typ: oUnion })
}

func (q *Query) Countall(s string) {
	q.push(&Op{ typ: oCountAll, name: s})
}

func (q *Query) Attr(a string) {
	q.Add(&Op{ typ: oAttr, name: a})		// split into name+value later.
}

func (q *Query) Pa() {
	if len(q.Stack) > 1 {	// XXX - horrible workaround so that the top element doesn't pop.
		q.Add(q.pop())
	}
}

func (q *Query) push(o *Op) {
	q.Stack = append(q.Stack, o)
}

func (q *Query) pop() *Op {
	l := len(q.Stack) - 1
	r := q.Stack[l]
	q.Stack = q.Stack[:l]
	return r
}

func (q *Query) Add(o *Op) {
	top := q.Stack[len(q.Stack)-1]
	top.contents = append(top.contents, o)
}

func (o Op) String() string {
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
