package grpc

import (
	userv1 "github.com/farid/user-service/api/proto/user/v1"
	"github.com/farid/user-service/internal/user/usecase"
)

type PaymentMethodHandler struct {
	uc *usecase.GetDefaultPaymentMethodUsecase
}

func NewPaymentMethodHandler(uc *usecase.GetDefaultPaymentMethodUsecase) *PaymentMethodHandler {
	return &PaymentMethodHandler{uc: uc}
}

func (h *PaymentMethodHandler) GetDefaultPaymentMethod(ctx context.Context, req *userv1.GetDefaultPaymentMethodRequest) (*userv1.PaymentMethodResponse, error) {
	pm, err := h.uc.Execute(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &userv1.PaymentMethodResponse{
		Type:  pm.Type,
		Last4: pm.Last4,
		Brand: pm.Brand,
	}, nil
}
