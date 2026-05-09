// Package sms provides a minimal interface for dispatching SMS messages.
// The production implementation should call the Telkomsel SMS API or a
// third-party aggregator (Vonage, Twilio). The stub implementation is safe
// for dev/staging and logs the message instead of sending it.
package sms

import "context"

// Client is the single method needed to dispatch an SMS.
type Client interface {
	Send(ctx context.Context, to, message string) error
}
