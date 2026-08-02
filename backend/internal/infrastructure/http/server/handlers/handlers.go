package handlers

import (
	"errors"
	"net/http"

	"github.com/chishkin-afk/intask/backend/pkg/errs"
)

type handlers struct {
	authService authService
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
		}
	}

	return http.StatusInternalServerError
}
