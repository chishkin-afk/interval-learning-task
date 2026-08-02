package handlers

import (
	"context"
	"net/http"

	"github.com/chishkin-afk/intask/backend/internal/application/dtos/requests"
	"github.com/chishkin-afk/intask/backend/internal/application/dtos/responses"
	"github.com/gin-gonic/gin"
)

type authService interface {
	Register(ctx context.Context, req *requests.AuthRequest) (*responses.Token, error)
	Login(ctx context.Context, req *requests.AuthRequest) (*responses.Token, error)
	GetSelf(ctx context.Context) (*responses.User, error)
	Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error)
	Delete(ctx context.Context) error
}

func (h *handlers) Register() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.AuthRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.Register(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.SetCookie(
			"token",
			resp.Token,
			int(resp.TTL.Seconds()),
			"", "", false, true)
		ctx.Status(http.StatusNoContent)
	}
}

func (h *handlers) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.AuthRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.Login(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.SetCookie(
			"token",
			resp.Token,
			int(resp.TTL.Seconds()),
			"", "", false, true)
		ctx.Status(http.StatusNoContent)
	}
}

func (h *handlers) GetSelf() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := h.authService.GetSelf(ctx.Request.Context())
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *handlers) UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpdateUser
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.Update(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *handlers) DeleteUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := h.authService.Delete(ctx.Request.Context()); err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
