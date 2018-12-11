package hexbytes

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type HexBytes []byte

func (hb HexBytes) MarshalJSON() ([]byte, error) {
	var st string
	if len(hb) > 0 {
		st = hexutil.Encode(hb)
	}
	return json.Marshal(st)
}

func (hb *HexBytes) UnmarshalJSON(data []byte) error {
	var st string
	err := json.Unmarshal(data, &st)
	if err != nil {
		return err
	}
	if len(st) == 0 {
		*hb = []byte{}
		return nil
	}
	b, err := hexutil.Decode(st)
	if err != nil {
		return err
	}
	*hb = b
	return nil
}
