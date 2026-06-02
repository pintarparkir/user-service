package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// metadataValue extracts a single metadata value by key from incoming gRPC context.
func metadataValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}
	return ""
}

// IdempotencyKeyFromCtx extracts the Idempotency-Key from gRPC metadata.
// Returns empty string if absent.
func IdempotencyKeyFromCtx(ctx context.Context) string {
	return metadataValue(ctx, HeaderIdempotencyKey)
}

// DriverIDFromCtx returns the driver id injected by the gateway after JWT verify.
func DriverIDFromCtx(ctx context.Context) string {
	return metadataValue(ctx, HeaderDriverID)
}

// CtxWithIdempotencyKey appends the key to outgoing metadata; useful for client-side calls.
func CtxWithIdempotencyKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, HeaderIdempotencyKey, key)
}
