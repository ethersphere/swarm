package sctx

import "context"

type (
	HTTPRequestIDKey struct{}
	requestHostKey   struct{}
	pushTagKey       struct{}
)

func SetHost(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, requestHostKey{}, domain)
}

func GetHost(ctx context.Context) string {
	v, ok := ctx.Value(requestHostKey{}).(string)
	if ok {
		return v
	}
	return ""
}

func SetPushTag(ctx context.Context, tag uint64) context.Context {
	return context.WithValue(ctx, pushTagKey{}, tag)
}

func GetPushTag(ctx context.Context) uint64 {
	v, ok := ctx.Value(pushTagKey{}).(uint64)
	if ok {
		return v
	}
	return 0
}
