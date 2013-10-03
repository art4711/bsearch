package index

import (
	"bytes"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

type blob_reader struct {
	file      *os.File
	mmap_data []byte
	mmap_size int
	Hdr       *IbHeader
}

func open_blob_reader(name string) (*blob_reader, error) {
	var br blob_reader
	var err error

	br.file, err = os.Open(name)
	if err != nil {
		return nil, err
	}
	s, err := br.file.Stat()
	if err != nil {
		br.file.Close()
		return nil, err
	}
	br.mmap_size = int(s.Size())
	br.mmap_data, err = syscall.Mmap(int(br.file.Fd()), 0, br.mmap_size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		br.file.Close()
		return nil, err
	}
	br.Hdr = br.get_header()
	return &br, nil
}

func (br *blob_reader) close() {
	syscall.Munmap(br.mmap_data)
	br.file.Close()
}

func (br *blob_reader) get_header() *IbHeader {
	return (*IbHeader)(unsafe.Pointer(&br.mmap_data[0]))
}

func (br *blob_reader) reslice(len uintptr, off, sz uint64, sz_in_bytes bool) unsafe.Pointer {
	if sz_in_bytes {
		sz /= uint64(len)
	}
	data := br.mmap_data[int(off) : int(off)+int(len)*int(sz)]
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	slice.Len /= int(len)
	slice.Cap = slice.Len

	return unsafe.Pointer(slice)
}

func (br *blob_reader) get_documents() []IbDocument {
	return *(*[]IbDocument)(br.reslice(unsafe.Sizeof(IbDocument{}), br.Hdr.documents_off, br.Hdr.ndocuments, false))
}

func (br *blob_reader) get_document_data(d *IbDocument) []byte {
	return br.mmap_data[d.Blob_offs : int(d.Blob_offs)+int(d.Doclen)-1]
}

func (br *blob_reader) get_invattrs() []IbInvattr {
	return *(*[]IbInvattr)(br.reslice(unsafe.Sizeof(IbInvattr{}), br.Hdr.invattrs_off, br.Hdr.ninvattrs, false))
}

func (br *blob_reader) get_attr_name(a *IbInvattr) string {
	s := br.mmap_data[a.Attr_offs:]
	return string(s[:bytes.IndexByte(s, 0)])
}

func (br *blob_reader) get_attr_docs(a *IbInvattr) []IbDoc {
	return *(*[]IbDoc)(br.reslice(unsafe.Sizeof(IbDoc{}), a.Docs_offs, a.Docslen, true))
}
