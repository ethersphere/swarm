package network

const (
	_ = iota
	META_ANNOUNCE_IPO
	META_ANNOUNCE_RENDERING_PERSISTENT
	META_ANNOUNCE_RENDERING_EPHEMERAL
	META_ANNOUNCE_MAX = META_ANNOUNCE_RENDERING_EPHEMERAL
)

type METAAnnounce struct {
	*METAEnvelope
	Payload []*METAPayload
}

func NewMETAAnnounce() (ma *METAAnnounce) {
	ma = &METAAnnounce{
		METAEnvelope: NewMETAEnvelope(),
	}
	return
}

func (ma *METAAnnounce) AddPayload(payloadtype uint8) error {
	
	payload := NewMETAPayload(payloadtype)
	if payload == nil {
		// return error
	}
	ma.Payload = append(ma.Payload, payload)
	return nil
}
