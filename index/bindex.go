// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package index

import (
	"bconf"
)

type Index struct {
	br     *blob_reader
	Docs   map[uint32][]byte
	Attrs  map[string][]IbDoc
	Meta   bconf.Bconf
	header string
}

func Open(name string) (*Index, error) {
	var in Index
	var err error

	in.br, err = open_blob_reader(name)
	if err != nil {
		return nil, err
	}

	in.Docs = make(map[uint32][]byte)
	for _, d := range in.br.get_documents() {
		in.Docs[d.Doc.Id] = in.br.get_document_data(&d)
	}

	in.Attrs = make(map[string][]IbDoc)
	for _, a := range in.br.get_invattrs() {
		in.Attrs[in.br.get_attr_name(&a)] = in.br.get_attr_docs(&a)
	}

	in.Meta.LoadJson(in.br.get_meta())

	in.Header() // Pre-cache the header to avoid race conditions.

	return &in, nil
}

func (in Index) Header() string {
	if in.header == "" {
		in.Meta.GetNode("attr", "order").ForeachSorted(func(k, v string) {
			if in.header != "" {
				in.header += "\t"
			}
			in.header += v
		})
	}
	return in.header
}

func (in Index) Close() {
	in.br.close()
}
