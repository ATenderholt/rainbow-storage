package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewChiMux(minio MinioHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Head("/*", minio.Proxy)
	r.Get("/*", minio.Proxy)
	r.Post("/*", minio.Proxy)
	r.Put("/*", minio.Proxy)
	r.Delete("/*", minio.Proxy)

	return r
}
