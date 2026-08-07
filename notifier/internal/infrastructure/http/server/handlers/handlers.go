package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/chishkin/intask/notifier/internal/application/dtos/requests"
	"github.com/chishkin/intask/notifier/internal/application/dtos/responses"
	"github.com/chishkin/intask/notifier/pkg/errs"
	"github.com/gin-gonic/gin"
)

type notifierService interface {
	SendMsg(ctx context.Context, req *requests.SendMsg) error
}

type handlers struct {
	notifierService notifierService
}

func (h *handlers) SendMsg() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.SendMsg
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: err.Error(),
			})
			return
		}

		if err := h.notifierService.SendMsg(ctx.Request.Context(), &req); err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

func (h *handlers) getCodeFromKind(err error) int {
	if errKind, ok := errors.AsType[*errs.KindError](err); ok {
		switch errKind.Kind() {
		case errs.KindRequest:
			return http.StatusBadRequest
		case errs.KindTimeout:
			return http.StatusRequestTimeout
		case errs.KindConflict:
			return http.StatusConflict
		case errs.KindNotFound:
			return http.StatusNotFound
		case errs.KindUnauth:
			return http.StatusUnauthorized
		case errs.KindPermissionDenied:
			return http.StatusForbidden
		}
	}

	return http.StatusInternalServerError
}
