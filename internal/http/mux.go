package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewChiMux(minio MinioHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	r.Post("/", minio.Proxy)

	return r
}
