package http

import (
	"context"

	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/sctx"
)

type contextKey int

const (
	uriKey contextKey = iota
	requestErrorKey
	requestCodeKey
	decryptedKey
	resolvedContentKey
)

func GetRUID(ctx context.Context) string {
	v, ok := ctx.Value(sctx.HTTPRequestIDKey).(string)
	if ok {
		return v
	}
	return "xxxxxxxx"
}

func SetRUID(ctx context.Context, ruid string) context.Context {
	return context.WithValue(ctx, sctx.HTTPRequestIDKey, ruid)
}

func GetURI(ctx context.Context) *api.URI {
	v, ok := ctx.Value(uriKey).(*api.URI)
	if ok {
		return v
	}
	return nil
}
func SetURI(ctx context.Context, uri *api.URI) context.Context {
	return context.WithValue(ctx, uriKey, uri)
}

// func GetEncrypted(ctx context.Context) bool {
// 	v, ok := ctx.Value(requestContextKey).(RequestContext)
// 	if ok {
// 		return v.Encrypted
// 	}
// 	return false
// }

func GetDecrypted(ctx context.Context) bool {
	v, ok := ctx.Value(decryptedKey).(bool)
	if ok {
		return v
	}
	return false
}
