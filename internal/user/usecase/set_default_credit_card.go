package usecase

import (
	"context"

	"github.com/farid/user-service/internal/user/repository"
)

type SetDefaultCreditCardUsecase struct {
	repo repository.CreditCardRepository
}

func NewSetDefaultCreditCardUsecase(repo repository.CreditCardRepository) *SetDefaultCreditCardUsecase {
	return &SetDefaultCreditCardUsecase{repo: repo}
}

func (uc *SetDefaultCreditCardUsecase) Execute(ctx context.Context, userID, cardID string) error {
	return uc.repo.SetDefaultCreditCard(ctx, userID, cardID)
}
