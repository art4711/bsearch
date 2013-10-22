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
	fmt.Printf("%*stype [%v]\n", indent, "", reflect.TypeOf(o));
	switch (o.(type)) {
	case *union:
		dumpUnion(o.(*union), indent + 1)
	case *intersection:
		dumpIntersection(o.(*intersection), indent + 1)
	}
}