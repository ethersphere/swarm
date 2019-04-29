package sctx

import "context"

type (
	HTTPRequestIDKey struct{}
	requestHostKey   struct{}
	tagKey           struct{}
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

func SetTag(ctx context.Context, tagId uint32) context.Context {
	return context.WithValue(ctx, tagKey{}, tagId)
}

func GetTag(ctx context.Context) uint32 {
	v, ok := ctx.Value(tagKey{}).(uint32)
	if ok {
		return v
	}
	return 0
}
