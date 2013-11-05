package parser_test

import (
	. "bsearch/parser"
	"testing"
)

func TestAll(t *testing.T) {
	s := ParseClassic("17 lim:10 count_all(hejsan) a:a b:a OR b:b").String()
	if s != `(offset [ 17 ] (limit [ 10 ] (count_all "hejsan" (intersection (attr "a:a") (union (attr "b:a") (attr "b:b"))))))` {
		t.Errorf("wrong result: %v", s)
	}
}