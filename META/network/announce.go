package network

const (
	_ = iota
	META_ANNOUNCE_IPO
	META_ANNOUNCE_RENDERING_PERSISTENT
	META_ANNOUNCE_RENDERING_EPHEMERAL
)

type METAAnnounce struct {
	METAHeader
	payload interface{}
}

func NewMETAAnnounce() (ma *METAAnnounce) {
	ma = &METAAnnounce{
		METAHeader: NewMETAEnvelope(),
	}
	return
}

