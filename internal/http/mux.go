package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

type RainbowContextKey string

const queriesContextKey = RainbowContextKey("queries")

func storeQueryKeys(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, request *http.Request) {
		var queries []string
		for key, _ := range request.URL.Query() {
			queries = append(queries, key)
		}

		ctx := context.WithValue(request.Context(), queriesContextKey, queries)
		r := request.Clone(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func getQueryKeys(request *http.Request) ([]string, bool) {
	ctx := request.Context()
	queries, ok := ctx.Value(queriesContextKey).([]string)
	return queries, ok
}

func NewChiMux(minio MinioHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, storeQueryKeys)

	// list buckets
	r.Get("/", minio.Proxy)

	r.Route("/{bucket}", func(r chi.Router) {
		r.Head("/*", minio.Proxy)

		r.With(minio.GetNotifications, minio.GetConfig).
			Get("/*", minio.Proxy)

		r.With(minio.SendNotifications).
			Post("/*", minio.Proxy)

		r.With(minio.PutNotifications, minio.SendNotifications, minio.PutConfig).
			Put("/*", minio.Proxy)

		r.Delete("/*", minio.Proxy)
	})

	return r
}
