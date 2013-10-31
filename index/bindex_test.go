// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package index_test

import (
	"bsearch/index"
	"testing"
)

func TestOpen(t *testing.T) {
	in, err := index.Open("/Users/art/db.blob")
	if err != nil {
		t.Fatalf("bindex.Open: %v\n", err)
	}
	defer in.Close()
	t.Log("%v\n", string(in.Docs[3]))

	for k, v := range in.Attrs {
		t.Logf("%v -> %v\n", k, v)
	}

	t.FailNow()
}
