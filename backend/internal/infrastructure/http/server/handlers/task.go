package handlers

import (
	"context"
	"net/http"

	"github.com/chishkin-afk/intask/backend/internal/application/dtos/requests"
	"github.com/chishkin-afk/intask/backend/internal/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type taskService interface {
	Save(ctx context.Context, req *requests.CreateTask) (*responses.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Task, error)
	ListAll(ctx context.Context, page, limit uint32) (*responses.ListTask, error)
	Update(ctx context.Context, taskID uuid.UUID, req *requests.UpdateTask) (*responses.Task, error)
	Delete(ctx context.Context, taskID uuid.UUID) error
}

func (h *handlers) CreateTask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.CreateTask
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.taskService.Save(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, resp)
	}
}

func (h *handlers) GetTaskByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: "invalid id of task",
			})
			return
		}

		resp, err := h.taskService.GetByID(ctx.Request.Context(), id)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *handlers) ListTasksAll() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page, limit := h.parsePagination(ctx)

		resp, err := h.taskService.ListAll(ctx.Request.Context(), page, limit)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *handlers) UpdateTask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpdateTask
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: err.Error(),
			})
			return
		}

		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: "invalid id of task",
			})
			return
		}

		resp, err := h.taskService.Update(ctx.Request.Context(), id, &req)
		if err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *handlers) DeleteTask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.Err{
				Error: "invalid id of task",
			})
			return
		}

		if err := h.taskService.Delete(ctx, id); err != nil {
			ctx.JSON(h.getCodeFromKind(err), &responses.Err{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
