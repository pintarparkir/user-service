package grpc

import (
	"context"
	pb "github.com/farid/user-service/pkg/pb/user/v1"
	"github.com/farid/user-service/internal/user/usecase"
)

type PaymentMethodHandler struct {
	uc *usecase.GetDefaultPaymentMethodUsecase
}

func NewPaymentMethodHandler(uc *usecase.GetDefaultPaymentMethodUsecase) *PaymentMethodHandler {
	return &PaymentMethodHandler{uc: uc}
}

func (h *PaymentMethodHandler) GetDefaultPaymentMethod(ctx context.Context, req *pb.GetDefaultPaymentMethodRequest) (*pb.PaymentMethod, error) {
	pm, err := h.uc.Execute(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.PaymentMethod{Type: pm.Type, CcToken: pm.CCToken, Last4: pm.Last4, Brand: pm.Brand}, nil
}
