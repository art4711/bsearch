package index

/*
 * This file contains the definitions of the on-disk format of an index blob.
 * The data structures here all have the Ib prefix (ib = index blob).
 */

// IbHeader is the header of the on-disk index. Contains file offsets to the other relevant data structures.
type IbHeader struct {
	Magic   uint64
	Version uint64

	ndocuments    uint64
	documents_off uint64

	ninvattrs    uint64
	invattrs_off uint64

	ninvwords      uint64
	invwords_off   uint64
	Total_word_len uint64

	meta_sz  uint64
	meta_off uint64
}

// IbDoc is the main key of a document.
// Id - unique document id.
// Order - main sorting order.
// Suborder - secondary sorting order.
type IbDoc struct {
	Id       uint32
	Order    uint32
	Suborder uint32
}

func (a *IbDoc) Cmp(b *IbDoc) int {
	if a.Order < b.Order {
		return -1
	}
	if a.Order > b.Order {
		return 1
	}
	if a.Id < b.Id {
		return -1
	}
	if a.Id > b.Id {
		return 1
	}
	return 0
}

func (a *IbDoc) Inc() *IbDoc {
	r := *a
	r.Id--
	return &r
}

// Returns a document higher than all possible documents in the index.
func NullDoc() *IbDoc {
	return &IbDoc{Id: ^uint32(0), Order: ^uint32(0)}
}

// IbDocument is the element of the documents array.
// Doc - document id.
// Doclen - length of the string representing the document in the blob.
// Blob_offs - offset in blob to the string.
type IbDocument struct {
	Doc       IbDoc
	Doclen    uint32
	Blob_offs uint64
}

// IbDocindex is the element of invwords array.
// Doc - document id.
// Posptr - offset into the per-word docpos array with the word positions for this document.
type IbDocindex struct {
	Doc    IbDoc
	Posptr uint32
}

type IbDocpos struct {
	Flags     uint16
	Pos       uint16
	Rel_boost uint16
}

type IbInvword struct {
	Docslen     uint64
	Word_offs   uint64
	Docs_offs   uint64
	Docops_offs uint64
}

type IbInvattr struct {
	Docslen   uint64
	Attr_offs uint64
	Docs_offs uint64
}
