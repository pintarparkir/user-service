package usecase

import (
	"context"

	"github.com/farid/user-service/internal/user/repository"
)

type PaymentMethod struct {
	Type    string
	CCToken string
	Last4   string
	Brand   string
}

type GetDefaultPaymentMethodUsecase struct {
	repo repository.CreditCardRepository
}

func NewGetDefaultPaymentMethodUsecase(repo repository.CreditCardRepository) *GetDefaultPaymentMethodUsecase {
	return &GetDefaultPaymentMethodUsecase{repo: repo}
}

func (uc *GetDefaultPaymentMethodUsecase) Execute(ctx context.Context, userID string) (*PaymentMethod, error) {
	card, err := uc.repo.GetDefaultCreditCard(ctx, userID)
	if err != nil {
		return &PaymentMethod{Type: "NONE"}, nil
	}
	if card == nil {
		return &PaymentMethod{Type: "NONE"}, nil
	}
	return &PaymentMethod{
		Type:    "CC",
		CCToken: card.Token,
		Last4:   card.Last4,
		Brand:   card.Brand,
	}, nil
}
