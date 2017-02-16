package network

const (
	_ = iota
	META_DATA_AUTHID
	META_DATA_WORK
	META_DATA_ARTIST
	META_DATA_MEDIA
	META_DATA_LICENCE
	META_DATA_USAGE
	META_DATA_MAX = META_DATA_USAGE
)

type METAPayload struct {
	Type uint8
	Label []string
	Data [][]byte
}

func NewMETAPayload(payloadtype uint8) *METAPayload {
	if payloadtype > META_DATA_MAX || payloadtype < META_CUSTOM {
		return nil
	}
	
	p := &METAPayload{
		Type: payloadtype,
	}
	
	return p
}
func (mtp *METAPayload) Add(label string, data []byte) error {
	mtp.Label = append(mtp.Label, label)
	mtp.Data = append(mtp.Data, data)
	return nil
}

func (mtp *METAPayload) GetType() uint8 {
	return mtp.Type
}

func (mtp *METAPayload) Length() int {
	return len(mtp.Label)
}

func (mtp *METAPayload) GetRawEntry(i int) (string, []byte) {
	if i >= mtp.Length() {
		return "", []byte{}
	}
	return mtp.Label[i], mtp.Data[i]
}
