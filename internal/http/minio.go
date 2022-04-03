package http

import (
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"net/http"
)

type MinioHandler struct {
	cfg *settings.Config
}

func NewMinioHandler(cfg *settings.Config) MinioHandler {
	return MinioHandler{
		cfg: cfg,
	}
}

func (h MinioHandler) Proxy(response http.ResponseWriter, request *http.Request) {

}
