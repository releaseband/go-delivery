package http

import "context"

type projectCtxKey struct{}

func SetProjectKey(ctx context.Context, projectKey string) context.Context {
	return context.WithValue(ctx, projectCtxKey{}, projectKey)
}

func GetProjectKey(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(projectCtxKey{}).(string)
	return val, ok
}
