package server

import "github.com/AndresCarvajalx/notiflow/logger"

func Start(port int) {

	logger.L.Sugar().Infof("Starting server on port %d", port)
}
