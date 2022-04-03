package settings

import (
	"github.com/ATenderholt/rainbow-storage/internal/logging"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = logging.NewLogger().Named("settings")
}
