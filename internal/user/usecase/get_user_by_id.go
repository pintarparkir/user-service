package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/farid/user-service/internal/user/model"
)

const userCacheTTL = 5 * time.Minute

// GetUserByID is the read-heavy path. We cache the marshalled user JSON in Redis
// for 5 minutes, with cache-miss falling back to Postgres.
//
// Cache invalidation: write paths (Update/Delete) DEL the key.
func (u *userUsecase) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	cacheKey := "user:" + id
	if u.cache != nil {
		if raw, err := u.cache.Get(ctx, cacheKey); err == nil && raw != "" {
			var cached model.User
			if jErr := json.Unmarshal([]byte(raw), &cached); jErr == nil {
				return &cached, nil
			}
		}
	}

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.cache != nil {
		if blob, mErr := json.Marshal(user); mErr == nil {
			_ = u.cache.Set(ctx, cacheKey, blob, userCacheTTL)
		}
	}
	return user, nil
}

// GetUserByExternalID — uncached (low frequency, used during sign-in flow).
func (u *userUsecase) GetUserByExternalID(ctx context.Context, externalUserID string) (*model.User, error) {
	return u.repo.GetByExternalID(ctx, externalUserID)
}
