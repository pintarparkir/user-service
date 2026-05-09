package usecase

import "context"

// DeleteUser is a soft delete (status='DELETED'). Idempotent — calling on a
// user that's already deleted returns nil so retries are safe.
func (u *userUsecase) DeleteUser(ctx context.Context, id string) error {
	if err := u.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	if u.cache != nil {
		_ = u.cache.Del(ctx, "user:"+id)
	}
	return nil
}
