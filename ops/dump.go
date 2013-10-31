// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package ops

import (
	"fmt"
	"reflect"
)

func dumpUnion(o *union, indent int) {
	for _, v := range *o {
		Dump(v, indent)
	}
}

func dumpIntersection(o *intersection, indent int) {
	for _, v := range *o {
		Dump(v, indent)
	}
}

func Dump(o interface{}, indent int) {
	fmt.Printf("%*stype [%v]\n", indent, "", reflect.TypeOf(o))
	switch o.(type) {
	case *union:
		dumpUnion(o.(*union), indent+1)
	case *intersection:
		dumpIntersection(o.(*intersection), indent+1)
	case *limit:
		Dump(o.(*limit).next, indent+1)
	case *offset:
		Dump(o.(*offset).next, indent+1)
	case *count_all:
		Dump(o.(*count_all).next, indent+1)
	}
}
