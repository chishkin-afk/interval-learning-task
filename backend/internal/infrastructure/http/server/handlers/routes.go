package handlers

import (
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

// New creates and configures a new Gin HTTP router.
//
// It initializes middleware, creates API route groups,
// and configures the router according to the current application environment.
//
// In development and local environments, the router uses Gin's default
// middleware stack with logger and recovery middleware.
// In production, only recovery middleware is enabled.
//
// Additional middleware can be provided through the mws parameter.
func New(
	cfg *config.Config,
	authService authService,
	taskService taskService,
	mws ...gin.HandlerFunc,
) *gin.Engine {
	router := getRouter(cfg)
	router.Use(mws...)

	handlers := handlers{
		authService: authService,
		taskService: taskService,
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			auth := v1.Group("/auth")
			{
				auth.POST("/register", handlers.Register())
				auth.POST("/login", handlers.Login())
				auth.GET("/user", handlers.GetSelf())
				auth.GET("/user/:id", handlers.GetByID())
				auth.PATCH("/user", handlers.UpdateUser())
				auth.DELETE("/user", handlers.DeleteUser())
			}

			task := v1.Group("/tasks")
			{
				task.POST("/task", handlers.CreateTask())
				task.GET("/task/:id", handlers.GetTaskByID())
				task.GET("/", handlers.ListTasksAll())
				task.PATCH("/task/:id", handlers.UpdateTask())
				task.DELETE("/task/:id", handlers.DeleteTask())
			}
		}
	}

	return router
}

// getRouter creates a Gin router configured for the current environment.
//
// Development and local environments use gin.Default(),
// which enables request logging and recovery middleware.
//
// Production environment uses a minimal router with recovery middleware only.
func getRouter(cfg *config.Config) *gin.Engine {
	if cfg.App.Env == config.EnvDev ||
		cfg.App.Env == config.EnvLocal {
		return gin.Default()
	}

	router := gin.New()
	router.Use(gin.Recovery())
	return router
}
