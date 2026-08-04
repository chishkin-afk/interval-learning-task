package authservice

import (
	"context"
	"errors"
	"log/slog"

	"github.com/chishkin-afk/intask/backend/internal/application/dtos/requests"
	"github.com/chishkin-afk/intask/backend/internal/application/dtos/responses"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	authctx "github.com/chishkin-afk/intask/backend/internal/infrastructure/context"
	"github.com/chishkin-afk/intask/backend/internal/modules/auth/domain/user"
	"github.com/chishkin-afk/intask/backend/pkg/errs"
	"github.com/google/uuid"
)

type JWTManager interface {
	Generate(userID uuid.UUID) (string, error)
	Validate(tokenString string) (uuid.UUID, error)
}

// AuthService provides authentication and authorization use cases.
type AuthService struct {
	cfg             *config.Config
	log             *slog.Logger
	userPersistence user.UserPersistenceRepository
	jwtMngr         JWTManager
}

// New creates a new AuthService instance.
//
// Returns a new AuthService instance.
func New(
	cfg *config.Config,
	log *slog.Logger,
	userPersistence user.UserPersistenceRepository,
	jwtMngr JWTManager,
) *AuthService {
	return &AuthService{
		cfg:             cfg,
		log:             log,
		userPersistence: userPersistence,
		jwtMngr:         jwtMngr,
	}
}

// Register creates a new user and returns a JWT token for the created account.
//
// Returns a validation error if the request is invalid or a conflict error
// if a user with the same email already exists.
func (as *AuthService) Register(ctx context.Context, req *requests.AuthRequest) (*responses.Token, error) {
	if req == nil {
		return nil, errs.NewKindError(errs.KindRequest, errs.ErrNilRequest)
	}

	user, err := as.newUser(req)
	if err != nil {
		return nil, err
	}

	token, err := as.newToken(user)
	if err != nil {
		return nil, err
	}

	if err := as.storeUser(ctx, user); err != nil {
		return nil, err
	}

	return token, nil
}

func (as *AuthService) storeUser(ctx context.Context, user *user.User) error {
	if err := as.userPersistence.Save(ctx, user); err != nil {
		as.log.Error("failed to store user",
			slog.String("operation", "store"),
			slog.String("error", err.Error()),
			slog.String("user_id", user.ID().String()),
		)

		return handleError(err)
	}

	return nil
}

func (as *AuthService) newUser(req *requests.AuthRequest) (*user.User, error) {
	email, err := user.NewEmail(req.Email)
	if err != nil {
		return nil, errs.NewKindError(errs.KindRequest, err)
	}

	user, err := user.New(email, req.Password)
	if err != nil {
		return nil, errs.NewKindError(errs.KindRequest, err)
	}

	return user, nil
}

// Login looking at existing user and returns a JWT token
//
// Returns a invalid credentials error if the password isn't equal or
// user by this email not found.
func (as *AuthService) Login(ctx context.Context, req *requests.AuthRequest) (*responses.Token, error) {
	email, err := user.NewEmail(req.Email)
	if err != nil {
		return nil, errs.NewKindError(errs.KindRequest, err)
	}

	user, err := as.getUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !user.PasswordHash().IsEqual(req.Password) {
		return nil, errs.NewKindError(errs.KindUnauth, errs.ErrInvalidCredentials)
	}

	return as.newToken(user)
}

func (as *AuthService) getUserByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	got, err := as.userPersistence.GetByEmail(ctx, email)
	if err != nil {
		as.log.Error("faield to get user by email",
			slog.String("operation", "get_by_id"),
			slog.String("error", err.Error()),
			slog.String("email", email.String()),
		)

		if errors.Is(err, user.ErrNotFound) {
			return nil, errs.NewKindError(errs.KindUnauth, errs.ErrInvalidCredentials)
		}

		return nil, handleError(err)
	}

	return got, nil
}

func (as *AuthService) newToken(user *user.User) (*responses.Token, error) {
	token, err := as.jwtMngr.Generate(user.ID())
	if err != nil {
		as.log.Error("failed to generate token",
			slog.String("operation", "generating_token"),
			slog.String("error", err.Error()),
			slog.String("user_id", user.ID().String()),
		)

		return nil, errs.NewKindError(errs.KindInternal, errs.ErrInternalServer)
	}

	return &responses.Token{
		Token: token,
		TTL:   as.cfg.JWT.TokenTTL,
	}, nil
}

// GetSelf returns an user by id
//
// It gets user from db by id and returns it as dto
func (as *AuthService) GetSelf(ctx context.Context) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := as.getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return as.userToDTO(user), nil
}

// GetSelf returns an user by id
//
// It gets user from db by id and returns it as dto
func (as *AuthService) GetByID(ctx context.Context, id uuid.UUID) (*responses.User, error) {
	user, err := as.getUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return as.userToDTO(user), nil
}

func (as *AuthService) getUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	got, err := as.userPersistence.GetByID(ctx, id)
	if err != nil {
		as.log.Error("faield to get user by email",
			slog.String("operation", "get_by_id"),
			slog.String("error", err.Error()),
			slog.String("user_id", id.String()),
		)

		if errors.Is(err, user.ErrNotFound) {
			return nil, errs.NewKindError(errs.KindUnauth, errs.ErrInvalidCredentials)
		}

		return nil, handleError(err)
	}

	return got, nil
}

// Update changes the value of someone fields of user
//
// Returns the updated user, or an error otherwise
func (as *AuthService) Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	updated, err := as.updateUser(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	return as.userToDTO(updated), nil
}

func (as *AuthService) updateUser(ctx context.Context, id uuid.UUID, req *requests.UpdateUser) (*user.User, error) {
	updated, err := as.userPersistence.Update(ctx, id, func(ctx context.Context, user *user.User) error {
		return as.applyUpdates(user, req)
	})
	if err != nil {
		as.log.Error("failed to update user",
			slog.String("operation", "update"),
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)

		return nil, handleError(err)
	}

	return updated, nil
}

func (as *AuthService) applyUpdates(user *user.User, req *requests.UpdateUser) error {
	if req.TgChatID != nil {
		user.ChangeTgChatID(*req.TgChatID)
	}
	if req.TgEnabled != nil {
		if *req.TgEnabled {
			return user.EnableTg()
		}

		return user.DisableTg()
	}

	return nil
}

// Delete removes user from store
//
// Returns an error if user isn't deleted or
// not found.
func (as *AuthService) Delete(ctx context.Context) error {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return err
	}

	return as.deleteUser(ctx, userID)
}

func (as *AuthService) deleteUser(ctx context.Context, id uuid.UUID) error {
	if err := as.userPersistence.Delete(ctx, id); err != nil {
		as.log.Error("failed to delete user",
			slog.String("operation", "delete"),
			slog.String("error", err.Error()),
			slog.String("user_id", id.String()),
		)

		return handleError(err)
	}

	return nil
}

func (as *AuthService) userToDTO(user *user.User) *responses.User {
	return &responses.User{
		ID:        user.ID(),
		Email:     user.Email().String(),
		TgChatID:  user.TgChatID(),
		TgEnabled: user.TgEnabled(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}
}

func (as *AuthService) getUserID(ctx context.Context) (uuid.UUID, error) {
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
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return errs.NewKindError(errs.KindTimeout, err)
	case errors.Is(err, user.ErrAlreadyExists):
		return errs.NewKindError(errs.KindConflict, err)
	case errors.Is(err, user.ErrNotFound):
		return errs.NewKindError(errs.KindNotFound, err)
	case errors.Is(err, user.ErrTgAlreadyDisabled),
		errors.Is(err, user.ErrTgAlreadyEnabled):
		return errs.NewKindError(errs.KindRequest, err)
	}

	return errs.NewKindError(errs.KindInternal, errs.ErrInternalServer)
}
