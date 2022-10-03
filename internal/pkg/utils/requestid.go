package utils

import "context"

type requestIDCtx string

const (
	requestIDCtxField requestIDCtx = "id"
)

func GetIDFromContext(ctx context.Context) string {
	id := ctx.Value(requestIDCtxField)
	if v, ok := id.(string); ok {
		return v
	}
	return ""
}

func GetCtxWithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDCtxField, id)
}
