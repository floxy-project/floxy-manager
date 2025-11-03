package middlewares

import (
	"net/http"

	"github.com/google/uuid"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
)

const (
	RequestIDHeader = "X-Request-Id"
)

func RequestIDMdw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		reqID := request.Header.Get(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := appcontext.WithRequestID(request.Context(), reqID)

		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
