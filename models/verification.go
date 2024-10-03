package models

import "time"

type VerificationDataType int

const (
	MailConfirmation VerificationDataType = iota + 1
	PassReset
)

// verification holds the verification info for a user creating their account
type Verification struct {
	Email     string               `json:"email"`
	Code      string               `json:"code"`
	ExpiresAt time.Time            `json:"expiresAt"`
	Type      VerificationDataType `json:"type"`
}
