package handlers

import (
	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, mws ...gin.HandlerFunc) *gin.Engine {
	router := getRouter(cfg)
	router.Use(mws...)

	// routes...

	return router
}

func getRouter(cfg *config.Config) *gin.Engine {
	if cfg.App.Env == config.EnvDev ||
		cfg.App.Env == config.EnvLocal {
		return gin.Default()
	}

	router := gin.New()
	router.Use(gin.Recovery())
	return router
}
