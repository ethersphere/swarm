package api

import (
	"context"
)

type contextKey string

const requestContextKey = contextKey("SwarmHTTPRequestContext")

type RequestContext struct {
	Uri       *URI
	Ruid      string // request unique id
	Encrypted bool   //indicates whether the request is flagged as encrypted
}

func SetRequestContext(ctx context.Context, requestContext RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, requestContext)
}

func GetRUID(ctx context.Context) string {
	v, ok := ctx.Value(requestContextKey).(RequestContext)
	if ok {
		return v.Ruid
	}
	return "ctxRuidErr"
}

func GetURI(ctx context.Context) *URI {
	v, ok := ctx.Value(requestContextKey).(RequestContext)
	if ok {
		return v.Uri
	}
	return nil
}

func GetEncrypted(ctx context.Context) bool {
	v, ok := ctx.Value(requestContextKey).(RequestContext)
	if ok {
		return v.Encrypted
	}
	return false
}
