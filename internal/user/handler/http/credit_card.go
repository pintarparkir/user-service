// Package http provides HTTP handlers for user-service, including credit card management.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/usecase"
)

// CreditCardHandler handles REST endpoints for credit card management.
type CreditCardHandler struct {
	addUC        *usecase.AddCreditCardUsecase
	setDefaultUC *usecase.SetDefaultCreditCardUsecase
	getMethodUC  *usecase.GetDefaultPaymentMethodUsecase
}

// NewCreditCardHandler creates a new CreditCardHandler.
func NewCreditCardHandler(addUC *usecase.AddCreditCardUsecase, setDefaultUC *usecase.SetDefaultCreditCardUsecase, getMethodUC *usecase.GetDefaultPaymentMethodUsecase) *CreditCardHandler {
	return &CreditCardHandler{
		addUC:        addUC,
		setDefaultUC: setDefaultUC,
		getMethodUC:  getMethodUC,
	}
}

// AddRequest represents a request to add a new credit card.
type AddRequest struct {
	CardNumber  string `json:"card_number" binding:"required"`
	ExpMonth    int    `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear     int    `json:"exp_year" binding:"required"`
	CVV         string `json:"cvv" binding:"required"`
	MakeDefault bool   `json:"make_default"`
}

// CardResponse represents the response after adding a credit card.
type CardResponse struct {
	ID        string `json:"id"`
	Last4     string `json:"last4"`
	Brand     string `json:"brand"`
	IsDefault bool   `json:"is_default"`
}

// SetDefaultRequest represents a request to set a card as default.
type SetDefaultRequest struct {
	CardID string `uri:"card_id" binding:"required"`
}

// PaymentMethodResponse represents the default payment method.
type PaymentMethodResponse struct {
	Type  string `json:"type"` // "CC", "QRIS", "NONE"
	Last4 string `json:"last4"`
	Brand string `json:"brand"`
}

// Add handles POST /v1/users/:user_id/credit-cards
// Creates a new credit card for the specified user.
func (h *CreditCardHandler) Add(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.addUC.Execute(c.Request.Context(), usecase.AddCreditCardRequest{
		UserID:      userID,
		CardNumber:  req.CardNumber,
		ExpMonth:    req.ExpMonth,
		ExpYear:     req.ExpYear,
		CVV:         req.CVV,
		MakeDefault: req.MakeDefault,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CardResponse{
		ID:        resp.ID,
		Last4:     resp.Last4,
		Brand:     resp.Brand,
		IsDefault: resp.IsDefault,
	})
}

// SetDefault handles PATCH /v1/users/:user_id/credit-cards/:card_id/default
// Sets the specified card as the default payment method for the user.
func (h *CreditCardHandler) SetDefault(c *gin.Context) {
	userID := c.Param("user_id")
	cardID := c.Param("card_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	if cardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "card_id is required"})
		return
	}

	if err := h.setDefaultUC.Execute(c.Request.Context(), userID, cardID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetDefault returns the user's default payment method (used internally by reservation-service via gRPC).
// This is a placeholder for documentation - actual implementation uses gRPC handler.
func (h *CreditCardHandler) GetDefault(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	pm, err := h.getMethodUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := PaymentMethodResponse{
		Type: pm.Type,
	}
	if pm.Type == "CC" {
		resp.Last4 = pm.Last4
		resp.Brand = pm.Brand
	}

	c.JSON(http.StatusOK, resp)
}
