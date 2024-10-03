package services

import (
	"gopkg.in/gomail.v2"
)

// TODO expand interface in future
type HCShowCalendarEmailService interface {
	CreateMail(mailReq *Mail) (*gomail.Message, error)
	SendMail(mailReq *Mail) error
	NewMail(from string, to []string, subject string, mailType MailType, data *MailData) *Mail
}

type MailType int

// List of Mail Types we are going to send.
const (
	MailConfirmation MailType = iota + 1
	PassReset
)

// MailData represents the data to be sent to the template of the mail.
// should this be expanded? in the future there could be different data that is sent through
type MailData struct {
	Email string
	Code  string
}

// Mail represents a email request
type Mail struct {
	from    string
	to      []string
	subject string
	mtype   MailType
	data    *MailData
}
