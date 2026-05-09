package model

import "time"

// User is the canonical profile aggregate.
//
// PII fields (PhoneE164, Email) live as plaintext in the domain object but are
// encrypted in the storage layer via pgcrypto. The repository handles the
// encrypt/decrypt round-trip so callers always work with plaintext.
type User struct {
	ID             string     `db:"id"`
	ExternalUserID string     `db:"external_user_id"`
	FullName       string     `db:"full_name"`
	PhoneE164      string     `db:"-"`               // decrypted by repository
	Email          string     `db:"-"`               // decrypted by repository
	PhoneE164Enc   []byte     `db:"phone_e164_enc"`  // bytea — encrypted at rest
	EmailEnc       []byte     `db:"email_enc"`       // bytea — encrypted at rest
	Status         UserStatus `db:"status"`
	Version        int        `db:"version"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}
