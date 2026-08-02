package middlewares

import (
	"context"
	"net/http"

	"github.com/chishkin-afk/intask/backend/internal/application/dtos/responses"
	authctx "github.com/chishkin-afk/intask/backend/internal/infrastructure/context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JWTManager interface {
	Validate(tokenString string) (uuid.UUID, error)
}

func NewAuthMiddleware(jwtMngr JWTManager, authRequire map[string]bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !authRequire[ctx.FullPath()] {
			ctx.Next()
			return
		}

		token, err := ctx.Cookie("token")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &responses.Err{
				Error: "invalid token",
			})
			return
		}

		userID, err := jwtMngr.Validate(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &responses.Err{
				Error: "invalid token",
			})
			return
		}

		tokenCtx := context.WithValue(ctx.Request.Context(), authctx.KeyUserID, userID)
		ctx.Request = ctx.Request.WithContext(tokenCtx)
		ctx.Next()
	}
}
