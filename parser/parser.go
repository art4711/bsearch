package parser

import (
	"strconv"
	"log"
	"bsearch/ops"
	"bsearch/index"
	"fmt"
)

type Query struct {
	Stack []ops.QueryContainer
	Offset uint32
	Limit uint32
	in *index.Index
}

func (q *Query) SetLimit(l string) {
	li, err := strconv.ParseInt(l, 10, 32)
	if err != nil {
		log.Fatal("limit parse failed")
	}
	q.Limit = uint32(li)
}

func (q *Query) SetOffset(o string) {
	oi, err := strconv.ParseInt(o, 10, 32)
	if err != nil {
		log.Fatal("limit parse failed")
	}
	q.Offset = uint32(oi)
}

func (q *Query) StartIntersection() {
	q.push(ops.NewIntersection())
}

func (q *Query) StartUnion() {
	q.push(ops.NewUnion())
}

func (q *Query) EndUnion() {
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
	q.Stack[len(q.Stack) - 1].Add(o)
}

func (q *Query) Attr(a string) {
	q.add(ops.NewAttr(q.in, a))
}

func (q *Query) SetIndex(i *index.Index) {
	q.in = i
}

func Parse(i *index.Index, s string) Query {
	q := &QueryParser{Buffer:s}

	q.SetIndex(i)
	q.Init()

	if err := q.Parse(); err != nil {
		log.Fatal(err)
	}
	q.Execute()
	fmt.Printf("%v\n", q.Query)
	return q.Query
}
