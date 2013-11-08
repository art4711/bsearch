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
	oInvalid optype = iota
	oAttr
	oUnion
	oIntersection
	oOffset
	oLimit
	oCountAll
)

var nameTyp = map[string]optype {
	"attr": oAttr,
	"union": oUnion,
	"intersection": oIntersection,
	"offset": oOffset,
	"limit": oLimit,
	"count_all": oCountAll,
}

type valtype int
const (
	vtNone valtype = iota
	vtInt
	vtString
	vtOp
)

type opattrs struct {
	name string
	valtyp valtype
	hasname, hascontents bool
	singlecontent bool
}

var opAttr = map[optype]opattrs{
	oInvalid: { name: "unknown" },
	oAttr: { name: "attr", hasname: true },
	oUnion: { name: "union", hascontents: true },
	oIntersection: { name: "intersection", hascontents: true },
	oOffset: { name: "offset", valtyp: vtInt, hascontents: true, singlecontent: true },
	oLimit: { name: "limit", valtyp: vtInt, hascontents: true, singlecontent: true },
	oCountAll: { name: "count_all", hasname: true, hascontents: true, singlecontent: true },
}

type Op struct {
	typ optype

	// counter name, attribute name, etc.
	name string

	// only one of the following can be set. counter attrs, attr range values, etc.
	strValue []string
	intValue []int64
	opValue []*Op

	contents []*Op
}

type Query struct {
	Stack []*Op
	Err []error
}

var ErrSyntax = errors.New("query syntax error")
var ErrLimitRange = errors.New("limit out of range")
var ErrOffsetRange = errors.New("offset out of range")
// Used in Generate if Parse error not handled
var ErrTyp = errors.New("invalid operation type")

func (q *Query) err(e error) {
	q.Err = append(q.Err, e)
}

func (q *Query) Lim(l string) {
	li, err := strconv.ParseInt(l, 10, 32)
	if err != nil {
		q.err(ErrLimitRange)
	}
	q.push(&Op{ typ: oLimit, intValue: []int64{ li } })
}

func (q *Query) Off(o string) {
	oi, err := strconv.ParseInt(o, 10, 32)
	if err != nil {
		q.err(ErrOffsetRange)
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

/*
 * Functions prefixed with Op* are the generalized way to add ops.
 */

func (q *Query) OpStart(typ string) {
	optyp, exists := nameTyp[typ]
	if !exists {
		q.err(errors.New(fmt.Sprintf("unknown operation type: [%v]", typ)))
		optyp = oInvalid
	}
	q.push(&Op{ typ: optyp })
}

func (q *Query) OpEnd() {
	top := q.pop()
	oa := opAttr[top.typ]

	// Validate the contents.
	if oa.hasname && top.name == "" {
		q.err(errors.New(fmt.Sprintf("Op %v needs a name and none provided", oa.name)))
	}
	if !oa.hasname && top.name != "" {
		q.err(errors.New(fmt.Sprintf("Op %v shouldn't have name: \"%v\"", oa.name, top.name)))
	}
	if oa.hascontents && len(top.contents) == 0 {
		q.err(errors.New(fmt.Sprintf("Op %v needs contents and none provided.", oa.name)))
	}
	if !oa.hascontents && len(top.contents) != 0 {
		q.err(errors.New(fmt.Sprintf("Op %v shouldn't have contents", oa.name)))
	}
	if oa.singlecontent && len(top.contents) != 1 {
		q.err(errors.New(fmt.Sprintf("Op %v should only have one content %d", oa.name, len(top.contents))))
	}
	if len(top.strValue) != 0 && oa.valtyp != vtString {
		q.err(errors.New(fmt.Sprintf("Op %v shouldn't have a strvalue: %v", oa.name, len(top.strValue))))
	}
	if len(top.intValue) != 0 && oa.valtyp != vtInt {
		q.err(errors.New(fmt.Sprintf("Op %v shouldn't have a intvalue: %v", oa.name, len(top.strValue))))
	}
	if len(top.opValue) != 0 && oa.valtyp != vtOp {
		q.err(errors.New(fmt.Sprintf("Op %v shouldn't have a opvalue: %v", oa.name, len(top.strValue))))
	}
	switch oa.valtyp {
	case vtInt:
		if len(top.intValue) == 0 {
			q.err(errors.New(fmt.Sprintf("Op %v needs intvalue", oa.name)))
		}
	case vtString:
		if len(top.strValue) == 0 {
			q.err(errors.New(fmt.Sprintf("Op %v needs strvalue", oa.name)))
		}
	case vtOp:
		if len(top.opValue) == 0 {
			q.err(errors.New(fmt.Sprintf("Op %v needs opvalue", oa.name)))
		}
	}
	if len(q.Stack) > 0 {	// XXX - horrible workaround so that the top element doesn't pop.
		q.Add(top)
	} else {
		q.push(top)
	}
}

func (q *Query) OpName(name string) {
	top := q.Stack[len(q.Stack)-1]
	top.name = name
}

func (q *Query) OpIntValue(val string) {
	top := q.Stack[len(q.Stack)-1]

	vi, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		q.err(err)
	}
	top.intValue = append(top.intValue, vi)
}

func (q *Query) OpStrValue(val string) {
	top := q.Stack[len(q.Stack)-1]
	top.strValue = append(top.strValue, val)
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
	for _, v := range o.strValue {
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
