// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package parser

import (
	"bsearch/index"
	"bsearch/ops"
)

func (o *op) generate(i *index.Index) (ops.QueryOp, error) {
	var qc ops.QueryContainer

	switch o.typ {
	case oAttr:
		return ops.NewAttr(i, o.name), nil
	case oUnion:
		qc = ops.NewUnion()
	case oIntersection:
		qc = ops.NewIntersection()
	case oOffset:
		qc = ops.NewOffset(uint(o.intValue[0]))
	case oLimit:
		qc = ops.NewLimit(uint(o.intValue[0]))
	case oCountAll:
		qc = ops.CountAll(o.name)
	}
	for _, v := range o.contents {
		c, err := v.generate(i)
		if err != nil {
			return nil, err
		}
		qc.Add(c)
	}
	return qc, nil
}

func Parse(i *index.Index, s string) (ops.QueryOp, error) {
	o, err := ParseQuery(s)
	if err != nil {
		return nil, err
	}
	return o.generate(i)
}
