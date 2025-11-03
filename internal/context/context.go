//nolint:revive // it's ok here
package context

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type contextKey string

const (
	ctxKeyProjectID  contextKey = "project_id"
	ctxKeyUserID     contextKey = "user_id"
	ctxKeyIsSuper    contextKey = "is_superuser"
	ctxKeyRawRequest contextKey = "raw_request"
	ctxKeyRequestID  contextKey = "request_id"
	ctxKeyUsername   contextKey = "username"
	ctxKeyParams     contextKey = "httprouter_params"
)

func WithProjectID(ctx context.Context, id domain.ProjectID) context.Context {
	return context.WithValue(ctx, ctxKeyProjectID, id)
}

func ProjectID(ctx context.Context) domain.ProjectID {
	return ctx.Value(ctxKeyProjectID).(domain.ProjectID) //nolint:forcetypeassert // ProjectID guaranteed
}

func WithUserID(ctx context.Context, userID domain.UserID) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func WithIsSuper(ctx context.Context, isSuper bool) context.Context {
	return context.WithValue(ctx, ctxKeyIsSuper, isSuper)
}

func IsSuper(ctx context.Context) bool {
	v, ok := ctx.Value(ctxKeyIsSuper).(bool)

	return ok && v
}

func UserID(ctx context.Context) domain.UserID {
	id, ok := ctx.Value(ctxKeyUserID).(domain.UserID)
	if !ok {
		return 0
	}

	return id
}

func WithRawRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, ctxKeyRawRequest, req)
}

func RawRequest(ctx context.Context) *http.Request {
	return ctx.Value(ctxKeyRawRequest).(*http.Request) //nolint:forcetypeassert // RawRequest guaranteed
}

func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, reqID)
}

func RequestID(ctx context.Context) string {
	v, ok := ctx.Value(ctxKeyRequestID).(string)
	if ok {
		return v
	}

	return ""
}

func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, ctxKeyUsername, username)
}

func Username(ctx context.Context) string {
	v, ok := ctx.Value(ctxKeyUsername).(string)
	if ok {
		return v
	}

	return ""
}

func WithParams(ctx context.Context, params httprouter.Params) context.Context {
	return context.WithValue(ctx, ctxKeyParams, params)
}

func Params(ctx context.Context) httprouter.Params {
	v, ok := ctx.Value(ctxKeyParams).(httprouter.Params)
	if !ok {
		return httprouter.Params{}
	}

	return v
}

func Param(ctx context.Context, name string) string {
	params := Params(ctx)
	return params.ByName(name)
}
