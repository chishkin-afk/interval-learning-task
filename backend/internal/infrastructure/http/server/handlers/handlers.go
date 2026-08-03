package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/chishkin-afk/intask/backend/pkg/errs"
	"github.com/gin-gonic/gin"
)

type handlers struct {
	authService authService
	taskService taskService
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

func (h *handlers) parsePagination(ctx *gin.Context) (uint32, uint32) {
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit < 1 {
		limit = 1
	}

	return uint32(page), uint32(limit)
}
