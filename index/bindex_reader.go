package index

import (
	"os"
	"unsafe"
	"github.com/art4711/filemap"
	"log"
)

type blob_reader struct {
	file      *os.File
	fmap	  *filemap.Map
	Hdr       *IbHeader
}

func open_blob_reader(name string) (*blob_reader, error) {
	var br blob_reader
	var err error

	br.file, err = os.Open(name)
	if err != nil {
		return nil, err
	}

	br.fmap, err = filemap.NewReader(br.file)
	if err != nil {
		br.file.Close()
		return nil, err
	}

	br.Hdr = br.get_header()
	return &br, nil
}

func (br *blob_reader) close() {
	br.fmap.Close()
	br.file.Close()
}

func (br *blob_reader) get_header() *IbHeader {
	return &(*(*[]IbHeader)(br.reslice(unsafe.Sizeof(IbHeader{}), 0, 1, false)))[0]
}

func (br *blob_reader) reslice(len uintptr, off, sz uint64, sz_in_bytes bool) unsafe.Pointer {
	if sz_in_bytes {
		sz /= uint64(len)
	}
	r, err := br.fmap.Slice(len, off, sz)
	if err != nil {
		log.Fatal("filemap.Slice: %v", err)
	} 
	return r
}

func (br *blob_reader) get_documents() []IbDocument {
	return *(*[]IbDocument)(br.reslice(unsafe.Sizeof(IbDocument{}), br.Hdr.documents_off, br.Hdr.ndocuments, false))
}

func (br *blob_reader) get_document_data(d *IbDocument) []byte {
	r, err := br.fmap.Bytes(d.Blob_offs, uint64(d.Doclen) - 1)
	if err != nil {
		log.Fatal("get_document_data: %v", err)
	}
	return r
}

func (br *blob_reader) get_invattrs() []IbInvattr {
	return *(*[]IbInvattr)(br.reslice(unsafe.Sizeof(IbInvattr{}), br.Hdr.invattrs_off, br.Hdr.ninvattrs, false))
}

func (br *blob_reader) get_attr_name(a *IbInvattr) string {
	r, err := br.fmap.CString(a.Attr_offs)
	if err != nil {
		log.Fatal("get_attr_name: %v", err)
	}
	return string(r)
}

func (br *blob_reader) get_attr_docs(a *IbInvattr) []IbDoc {
	return *(*[]IbDoc)(br.reslice(unsafe.Sizeof(IbDoc{}), a.Docs_offs, a.Docslen, true))
}

func (br *blob_reader) get_meta() []byte {
	r, err := br.fmap.Bytes(br.Hdr.meta_off, br.Hdr.meta_sz)
	if err != nil {
		log.Fatal("get_meta: %v", err)
	}
	return r
}
