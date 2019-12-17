package forward

var (
	sessionId = 0
)

type ForwardPeer struct {
}

type SessionInterface interface {
	Subscribe() <-chan ForwardPeer
	Get(numberOfPeers int) ([]ForwardPeer, error)
	Close()
}

// also implements context.Context
type SessionContext struct {
	CapabilityIndex string
	SessionId       int
}

func NewSessionContext(cpidx string) SessionContext {
	sctx := SessionContext{
		CapabilityIndex: cpidx,
		SessionId:       sessionId,
	}
	sessionId++
	return sctx
}
