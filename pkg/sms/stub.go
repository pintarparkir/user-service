package sms

import (
	"context"

	"github.com/farid/user-service/pkg/logger"
)

type stubClient struct{}

// NewStubClient returns a no-op SMS client that logs instead of calling an API.
// Swap for a real implementation when the Telkomsel SMS gateway is available.
func NewStubClient() Client { return &stubClient{} }

func (s *stubClient) Send(ctx context.Context, to, message string) error {
	logger.Info(ctx, "SMS dispatched (stub)", map[string]interface{}{
		"to":      to,
		"message": message,
	})
	return nil
}
