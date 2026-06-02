// Package usecase implements credit card business logic.
package usecase

import (
	"context"
	"strings"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/internal/user/repository"
)

// ── Add Credit Card ─────────────────────────────────────────────────────────

type AddCreditCardRequest struct {
	UserID      string
	CardNumber  string
	ExpMonth    int
	ExpYear     int
	CVV         string
	MakeDefault bool
}

type AddCreditCardResponse struct {
	ID        string
	Last4     string
	Brand     string
	IsDefault bool
}

type AddCreditCardUsecase struct {
	repo repository.CreditCardRepository
}

func NewAddCreditCardUsecase(repo repository.CreditCardRepository) *AddCreditCardUsecase {
	return &AddCreditCardUsecase{repo: repo}
}

func (uc *AddCreditCardUsecase) Execute(ctx context.Context, req AddCreditCardRequest) (*AddCreditCardResponse, error) {
	last4 := req.CardNumber[len(req.CardNumber)-4:]
	brand := detectBrand(req.CardNumber)
	token := "tok_" + last4

	card := &model.CreditCard{
		UserID:    req.UserID,
		Token:     token,
		Last4:     last4,
		Brand:     brand,
		IsDefault: req.MakeDefault,
	}
	if err := uc.repo.AddCreditCard(ctx, card); err != nil {
		return nil, err
	}
	return &AddCreditCardResponse{ID: card.ID, Last4: card.Last4, Brand: card.Brand, IsDefault: card.IsDefault}, nil
}

func detectBrand(cardNumber string) string {
	if strings.HasPrefix(cardNumber, "4") {
		return "VISA"
	}
	if strings.HasPrefix(cardNumber, "5") {
		return "MASTERCARD"
	}
	return "UNKNOWN"
}

// ── Set Default Credit Card ─────────────────────────────────────────────────

type SetDefaultCreditCardUsecase struct {
	repo repository.CreditCardRepository
}

func NewSetDefaultCreditCardUsecase(repo repository.CreditCardRepository) *SetDefaultCreditCardUsecase {
	return &SetDefaultCreditCardUsecase{repo: repo}
}

func (uc *SetDefaultCreditCardUsecase) Execute(ctx context.Context, userID, cardID string) error {
	return uc.repo.SetDefaultCreditCard(ctx, userID, cardID)
}

// ── Get Default Payment Method ──────────────────────────────────────────────

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
	if err != nil || card == nil {
		return &PaymentMethod{Type: "NONE"}, nil
	}
	return &PaymentMethod{
		Type:    "CC",
		CCToken: card.Token,
		Last4:   card.Last4,
		Brand:   card.Brand,
	}, nil
}
