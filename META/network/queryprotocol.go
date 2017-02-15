package network

const (
	_ = iota
	META_QUERY_REQUEST
	META_QUERY_REPLY
)

type METAQuery struct {
	METAHeader
	payload interface{}
}

func NewMETAQuery() (ma *METAQuery) {
	ma = &METAQuery{
		METAHeader: NewMETAEnvelope(),
	}
	return
}

