// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package parser_test

import (
	. "bsearch/parser"
	"testing"
	"github.com/art4711/timers"
)

func TestClassic(t *testing.T) {
	ops, _ := ParseClassic("17 lim:10 count_all(hejsan) a:a b:a OR b:b")
	s := ops.String()
	if s != `(offset [ 17 ] (limit [ 10 ] (count_all "hejsan" (intersection (attr "a:a") (union (attr "b:a") (attr "b:b"))))))` {
		t.Errorf("wrong result: %v", s)
	}
}

func TestStructured(t *testing.T) {
	q := `(offset [ 17 ] (limit [ 10 ] (count_all "hejsan" (intersection (attr "a:a") (union (attr "b:a") (attr "b:b"))))))`
	ops, err := ParseStructured(q, timers.New().Start("hej"))
	if err != nil {
		t.Fatal(err)
	}
	s := ops.String()
	if s != q {
		t.Errorf("wrong result: %v", s)
	}
}

func BenchmarkStructured(b *testing.B) {
	q := `(offset [ 17 ] (limit [ 10 ] (count_all "hejsan" (intersection (attr "a:a") (union (attr "b:a") (attr "b:b"))))))`
	for i:= 0; i < b.N; i++ {
		_, err := ParseStructured(q, timers.New().Start("hej"))
		if err != nil {
			b.Fatal(err)
		}
	}
}