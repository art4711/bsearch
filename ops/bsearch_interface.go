package ops

import (
	"bsearch/index"
)

type QueryOp interface {
	CurrentDoc() *index.IbDoc
	NextDoc(*index.IbDoc) *index.IbDoc
}
