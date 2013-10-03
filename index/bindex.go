package index

type Index struct {
	br    *blob_reader
	Docs  map[uint32][]byte
	Attrs map[string][]IbDoc
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
	return &in, nil
}

func (in *Index) Close() {
	in.br.close()
}
