package userredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/chishkin-afk/intask/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type userRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	cache Cache,
) *userRepository {
	return &userRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (ur *userRepository) SetByCode(ctx context.Context, uid uuid.UUID, code int) error {
	ur.log.Debug("setting user by code",
		slog.String("user_id", uid.String()),
	)

	if err := ur.setByCode(ctx, ur.cache, uid, code); err != nil {
		return fmt.Errorf("can't set by code: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) setByCode(ctx context.Context, cache Cache, uid uuid.UUID, code int) error {
	bytes, err := uidToBytes(uid)
	if err != nil {
		return err
	}

	key := getCodeKey(code)
	return cache.Set(ctx, key, bytes, ur.cfg.Cache.CodeTTL).Err()
}

func (ur *userRepository) DelByCode(ctx context.Context, code int) error {
	ur.log.Debug("deleting user by code")

	if err := ur.delByCode(ctx, ur.cache, code); err != nil {
		return fmt.Errorf("can't del by code: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) delByCode(ctx context.Context, cache Cache, code int) error {
	key := getCodeKey(code)
	return cache.Del(ctx, key).Err()
}

func (ur *userRepository) GetByCode(ctx context.Context, code int) (uuid.UUID, error) {
	ur.log.Debug("getting user by code")

	uid, err := ur.getByCode(ctx, ur.cache, code)
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't get user by code: %w",
			handleError(err),
		)
	}

	return uid, nil
}

func (ur *userRepository) getByCode(ctx context.Context, cache Cache, code int) (uuid.UUID, error) {
	bytes, err := cache.Get(ctx, getCodeKey(code)).Bytes()
	if err != nil {
		return uuid.Nil, err
	}

	return bytesToUid(bytes)
}

func getCodeKey(code int) string {
	return fmt.Sprintf("code: %d", code)
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return user.ErrNotFound
	}

	return err
}
