package taskservice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/chishkin-afk/intask/backend/internal/application/dtos/requests"
	"github.com/chishkin-afk/intask/backend/internal/application/dtos/responses"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	authctx "github.com/chishkin-afk/intask/backend/internal/infrastructure/context"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/workerpool"
	"github.com/chishkin-afk/intask/backend/internal/modules/task/domain/task"
	"github.com/chishkin-afk/intask/backend/pkg/errs"
	"github.com/google/uuid"
)

type authService interface {
}

type workerPool interface {
	Submit(ctx context.Context, fn func(context.Context) error) error
}

// TaskService provides application-level operations for managing tasks.
//
// It orchestrates the creation, validation, and persistence of tasks,
// acting as a bridge between the delivery layer (HTTP/gRPC) and the domain model.
type TaskService struct {
	cfg             *config.Config
	log             *slog.Logger
	taskPersistence task.TaskPersistenceRepository
	wp              workerPool
}

// New creates and returns a new instance of TaskService.
//
// It requires a structured logger for operational visibility and a repository
// implementation for task persistence.
func New(
	cfg *config.Config,
	log *slog.Logger,
	taskPersistence task.TaskPersistenceRepository,
	wp workerPool,
) *TaskService {
	return &TaskService{
		cfg:             cfg,
		log:             log,
		taskPersistence: taskPersistence,
		wp:              wp,
	}
}

// Save creates a new task from the given request and stores it in the database.
//
// It relies on the context to extract the authenticated user's ID, ensuring
// the task is securely associated with the correct owner.
//
// Returns a task response DTO on success. If the request is invalid, the user
// is unauthenticated, or a persistence error occurs, an appropriate domain error is returned.
func (ts *TaskService) Save(ctx context.Context, req *requests.CreateTask) (*responses.Task, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	task, err := ts.newTask(userID, req)
	if err != nil {
		return nil, err
	}

	if err := ts.storeTask(ctx, task); err != nil {
		return nil, err
	}

	return taskToDTO(task), nil
}

func (ts *TaskService) storeTask(ctx context.Context, task *task.Task) error {
	if err := ts.taskPersistence.Save(ctx, task); err != nil {
		ts.log.Error("can't store task",
			slog.String("operation", "store"),
			slog.String("user_id", task.UserID().String()),
			slog.String("task_id", task.ID().String()),
			slog.String("error", err.Error()),
		)

		return handleError(err)
	}

	return nil
}

func (ts *TaskService) newTask(userID uuid.UUID, req *requests.CreateTask) (*task.Task, error) {
	task, err := task.New(
		userID,
		req.Title,
		task.LeetcodeURL(req.LeetcodeURL),
	)
	if err != nil {
		return nil, errs.NewKindError(errs.KindRequest, err)
	}

	return task, nil
}

// GetByID retrieves a task by its unique identifier and enforces strict ownership validation.
//
// It extracts the authenticated user from the context and verifies that the requested
// task belongs to them. This prevents Insecure Direct Object Reference (IDOR) vulnerabilities
// by ensuring users can only access their own data.
//
// Returns a task response DTO upon success. Potential errors include unauthenticated
// requests (missing context), not found (task does not exist), permission denied
// (user is not the author), or internal persistence failures.
func (ts *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Task, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	task, err := ts.getTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task.UserID() != userID {
		return nil, errs.NewKindError(errs.KindPermissionDenied, errs.ErrUserNotAuthor)
	}

	return taskToDTO(task), nil
}

func (ts *TaskService) getTaskByID(ctx context.Context, id uuid.UUID) (*task.Task, error) {
	task, err := ts.taskPersistence.GetByID(ctx, id)
	if err != nil {
		ts.log.Error("can't get task by id",
			slog.String("operation", "get_by_id"),
			slog.String("error", err.Error()),
			slog.String("task_id", id.String()),
		)

		return nil, handleError(err)
	}

	return task, nil
}

// ListAll retrieves a paginated list of tasks for the authenticated user.
//
// It enforces strict data isolation by querying only the tasks owned by the user
// extracted from the context. The method validates pagination parameters to prevent
// excessive database load and returns metadata for cursor/page-based navigation.
//
// Returns a list task response DTO containing the items and pagination info.
// Potential errors include invalid pagination parameters, unauthenticated requests,
// or internal persistence failures.
func (ts *TaskService) ListAll(ctx context.Context, page, limit uint32) (*responses.ListTask, error) {
	if err := validatePagination(page, limit); err != nil {
		return nil, err
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	list, count, err := ts.listAll(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return listToDTO(list, count, page, limit), nil
}

func (ts *TaskService) listAll(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*task.Task, int64, error) {
	list, count, err := ts.taskPersistence.ListAll(ctx, userID, page, limit)
	if err != nil {
		ts.log.Error("can't list all the tasks",
			slog.String("operation", "list_all"),
			slog.String("error", err.Error()),
			slog.Int("page", int(page)),
			slog.Int("limit", int(limit)),
		)

		return nil, 0, handleError(err)
	}

	return list, count, nil
}

// Update modifies an existing task identified by its UUID using the provided request payload.
//
// It employs a transactional callback pattern to safely apply partial updates while
// enforcing strict ownership validation. Only fields explicitly set in the request
// are modified; nil fields are ignored.
//
// Returns the updated task response DTO upon success. Potential errors include
// unauthenticated requests, not found (task does not exist), permission denied
// (user is not the author), invalid field values, or internal persistence failures.
func (ts *TaskService) Update(ctx context.Context, taskID uuid.UUID, req *requests.UpdateTask) (*responses.Task, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	task, err := ts.updateTask(ctx, userID, taskID, req)
	if err != nil {
		return nil, err
	}

	return taskToDTO(task), nil
}

func (ts *TaskService) updateTask(ctx context.Context, userID, taskID uuid.UUID, req *requests.UpdateTask) (*task.Task, error) {
	updFunc := func(ctx context.Context, task *task.Task) error {
		if task.UserID() != userID {
			return errs.ErrUserNotAuthor
		}

		return ts.applyChanges(task, req)
	}

	task, err := ts.taskPersistence.Update(ctx, taskID, updFunc)
	if err != nil {
		ts.log.Error("can't update task",
			slog.String("operation", "update"),
			slog.String("task_id", taskID.String()),
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)

		return nil, handleError(err)
	}

	return task, nil
}

func (ts *TaskService) applyChanges(updTask *task.Task, req *requests.UpdateTask) error {
	if req.Title != nil {
		if err := updTask.ChangeTitle(*req.Title); err != nil {
			return err
		}
	}

	if req.LeetcodeURL != nil {
		return updTask.ChangeLeetcodeURL(task.LeetcodeURL(*req.LeetcodeURL))
	}

	return nil
}

// Delete permanently removes a task identified by its UUID.
//
// It enforces strict ownership validation by first retrieving the task and verifying
// that the authenticated user from the context is the author. This prevents
// unauthorized deletion of other users' data.
//
// Returns nil on successful deletion. Potential errors include unauthenticated
// requests, not found (task does not exist), permission denied (user is not the author),
// or internal persistence failures.
func (ts *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	task, err := ts.getTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task.UserID() != userID {
		return errs.NewKindError(errs.KindPermissionDenied, errs.ErrUserNotAuthor)
	}

	return ts.deleteTask(ctx, task.ID())
}

func (ts *TaskService) deleteTask(ctx context.Context, taskID uuid.UUID) error {
	if err := ts.taskPersistence.Delete(ctx, taskID); err != nil {
		ts.log.Error("can't delete task",
			slog.String("operation", "delete"),
			slog.String("error", err.Error()),
			slog.String("task_id", taskID.String()),
		)

		return handleError(err)
	}

	return nil
}

// RunNotificator starts a background loop that periodically submits
// notification tasks to the worker pool based on the configured ticker interval.
//
// It respects context cancellation for graceful shutdown and handles
// worker pool backpressure by skipping ticks when the pool is saturated.
// This method blocks until the context is canceled or the pool is stopped.
func (ts *TaskService) RunNotificator(ctx context.Context) {
	ticker := time.NewTicker(ts.cfg.Service.TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ts.log.Error("notificator was stopped by ctx",
				slog.String("error", ctx.Err().Error()),
			)

			return
		case <-ticker.C:
			if err := ts.wp.Submit(ctx, ts.notificate); err != nil {
				if errors.Is(err, workerpool.ErrPoolIsDone) ||
					errors.Is(err, workerpool.ErrPoolIsStop) {
					return
				}

				ts.log.Warn("can't submit task into worker pool",
					slog.String("error", err.Error()),
				)
			}
		}
	}
}

func (ts *TaskService) notificate(ctx context.Context) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, ts.cfg.Service.NotificateTimeout)
	defer cancel()

	list, err := ts.listByNotification(ctxTimeout)
	if err != nil {
		ts.log.Error("can't list tasks by notification",
			slog.String("error", err.Error()),
		)

		return err
	}

	_ = list

	// TODO: сделать получения всех tg chat id по user id тасок
	// 		 и сделать фейковый (пока что) вызов notifier

	return nil
}

func (ts *TaskService) listByNotification(ctx context.Context) ([]*task.Task, error) {
	list, _, err := ts.taskPersistence.ListByNotification(ctx)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func validatePagination(page, limit uint32) error {
	if page < 1 || page > 100 {
		return errs.NewKindError(errs.KindRequest, errs.ErrInvalidPage)
	}

	if limit < 1 || limit > 1000 {
		return errs.NewKindError(errs.KindRequest, errs.ErrInvalidLimit)
	}

	return nil
}

func listToDTO(list []*task.Task, count int64, page, limit uint32) *responses.ListTask {
	totalPages := count + (int64(limit)-1)/int64(limit)
	resp := &responses.ListTask{
		List:        make([]*responses.Task, len(list)),
		TotalPages:  totalPages,
		HasNextPage: totalPages > int64(page),
		HasPrevPage: page > 1,
	}

	for i, task := range list {
		resp.List[i] = taskToDTO(task)
	}

	return resp
}

func taskToDTO(task *task.Task) *responses.Task {
	return &responses.Task{
		ID:          task.ID(),
		UserID:      task.UserID(),
		Title:       task.Title(),
		LeetcodeURL: task.LeetcodeURL().String(),
		NextNotify:  task.NextNotify(),
		NotifyCount: task.NotifyCount(),
		IsActive:    task.IsActive(),
		CreatedAt:   task.CreatedAt(),
	}
}

func getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(authctx.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewKindError(errs.KindUnauth, errs.ErrInvalidToken)
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.NewKindError(errs.KindUnauth, errs.ErrInvalidToken)
}

func handleError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewKindError(errs.KindTimeout, err)
	case errors.Is(err, task.ErrNotFound):
		return errs.NewKindError(errs.KindNotFound, err)
	case errors.Is(err, task.ErrUserNotFound):
		return errs.NewKindError(errs.KindNotFound, err)
	case errors.Is(err, task.ErrInvalidTitle),
		errors.Is(err, task.ErrInvalidLeetcodeURL):
		return errs.NewKindError(errs.KindRequest, err)
	case errors.Is(err, errs.ErrUserNotAuthor):
		return errs.NewKindError(errs.KindPermissionDenied, err)
	}

	return errs.NewKindError(errs.KindInternal, errs.ErrInternalServer)
}
