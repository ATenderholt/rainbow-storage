package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewChiMux(minio MinioHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// list buckets
	r.Get("/", minio.Proxy)

	r.Route("/{bucket}", func(r chi.Router) {
		r.Head("/*", minio.Proxy)

		r.With(minio.GetNotifications, minio.GetTransferAcceleration).
			Get("/*", minio.Proxy)

		r.With(minio.SendNotifications).
			Post("/*", minio.Proxy)

		r.With(minio.PutNotifications, minio.SendNotifications, minio.PutTransferAcceleration).
			Put("/*", minio.Proxy)

		r.Delete("/*", minio.Proxy)
	})

	return r
}
