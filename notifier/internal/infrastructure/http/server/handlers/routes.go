package handlers

import (
	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, ns notifierService, mws ...gin.HandlerFunc) *gin.Engine {
	router := getRouter(cfg)
	router.Use(mws...)

	handlers := handlers{
		notifierService: ns,
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/send", handlers.SendMsg())
		}
	}

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
