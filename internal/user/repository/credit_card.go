package repository

import (
	"context"
	"github.com/farid/user-service/internal/user/model"
)

type CreditCardRepository interface {
	AddCreditCard(ctx context.Context, card *model.CreditCard) error
	GetCreditCardsByUserID(ctx context.Context, userID string) ([]model.CreditCard, error)
	GetDefaultCreditCard(ctx context.Context, userID string) (*model.CreditCard, error)
	SetDefaultCreditCard(ctx context.Context, userID, cardID string) error
	DeleteCreditCard(ctx context.Context, cardID string) error
}
