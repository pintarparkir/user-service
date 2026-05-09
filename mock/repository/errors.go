package repository

import apperror "github.com/farid/user-service/pkg/error"

func notFound() error { return apperror.ErrNotFound }
func conflict() error { return apperror.ErrConflict }
