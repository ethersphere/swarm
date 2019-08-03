package protocols

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/spancontext"
	opentracing "github.com/opentracing/opentracing-go"
)

// msgWithContext is used to propagate marshalled context alongside message payloads
type msgWithContext struct {
	Context []byte
	Msg     []byte
}

func encodeWithContext(ctx context.Context, msg interface{}) (interface{}, int, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	tracer := opentracing.GlobalTracer()
	sctx := spancontext.FromContext(ctx)
	if sctx != nil {
		err := tracer.Inject(
			sctx,
			opentracing.Binary,
			writer)
		if err != nil {
			return nil, 0, err
		}
	}
	writer.Flush()
	msgBytes, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return nil, 0, err
	}

	return &msgWithContext{
		Context: b.Bytes(),
		Msg:     msgBytes,
	}, len(msgBytes), nil
}

func decodeWithContext(msg p2p.Msg) (context.Context, []byte, error) {
	var wmsg msgWithContext
	err := msg.Decode(&wmsg)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	if len(wmsg.Context) == 0 {
		return ctx, wmsg.Msg, nil
	}

	tracer := opentracing.GlobalTracer()
	sctx, err := tracer.Extract(opentracing.Binary, bytes.NewReader(wmsg.Context))
	if err != nil {
		return nil, nil, err
	}
	ctx = spancontext.WithContext(ctx, sctx)
	return ctx, wmsg.Msg, nil
}

func encodeWithoutContext(ctx context.Context, msg interface{}) (interface{}, int, error) {
	return msg, 0, nil
}

func decodeWithoutContext(msg p2p.Msg) (context.Context, []byte, error) {
	b, err := ioutil.ReadAll(msg.Payload)
	if err != nil {
		return nil, nil, err
	}
	return context.Background(), b, nil
}
