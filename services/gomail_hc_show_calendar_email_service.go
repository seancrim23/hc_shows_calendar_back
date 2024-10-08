package services

import (
	"bytes"
	"errors"
	"fmt"
	"hc_shows_calendar_back/utils"
	"os"
	"text/template"

	"gopkg.in/gomail.v2"
)

// gomail implementation of email service interface
type GomailHcShowCalendarEmailService struct{}

// tbd if anything needs to be injected into this service on creation
func NewGomailHcShowCalendarEmailService() *GomailHcShowCalendarEmailService {
	return &GomailHcShowCalendarEmailService{}
}

func (g *GomailHcShowCalendarEmailService) CreateMail(mailReq *Mail) (*gomail.Message, error) {
	m := gomail.NewMessage()
	m.SetHeader(utils.FROM, mailReq.from)
	m.SetHeader(utils.TO, mailReq.to...)
	m.SetHeader(utils.SUBJECT, mailReq.subject)

	var emailBody = ""
	var err error
	switch mailReq.mtype {
	case MailConfirmation:
		emailBody, err = buildEmailBody(mailTemplateLocations[MailConfirmation], *mailReq.data)
	case PassReset:
		emailBody, err = buildEmailBody(mailTemplateLocations[PassReset], *mailReq.data)
	default:
		fmt.Println("email type is required")
		return nil, errors.New("invalid email type")
	}

	if err != nil {
		fmt.Println("error creating mail")
		return nil, err
	}

	m.SetBody(utils.TEXT_HTML, emailBody)

	return m, nil
}

func (g *GomailHcShowCalendarEmailService) SendMail(mailReq *Mail) error {
	m, err := g.CreateMail(mailReq)
	if err != nil {
		fmt.Println("error creating mail")
		fmt.Println(err)
		return err
	}

	d := gomail.NewDialer(utils.GMAIL_SMTP, utils.GMAIL_SMTP_PORT, utils.MY_EMAIL, os.Getenv(utils.GMAIL_AUTH))
	err = d.DialAndSend(m)
	if err != nil {
		fmt.Println("error dialing and sending")
		fmt.Println(err)
		return err
	}

	return nil
}

func (g *GomailHcShowCalendarEmailService) NewMail(from string, to []string, subject string, mailType MailType, data *MailData) *Mail {
	return &Mail{
		from:    from,
		to:      to,
		subject: subject,
		mtype:   mailType,
		data:    data,
	}
}

func buildEmailBody(templatePath string, mailData MailData) (string, error) {
	var body bytes.Buffer

	//TODO this depends on cloud storage... should be able to have a lower level solution too
	fs, err := utils.ReadFileFromBucket(templatePath, os.Getenv(utils.TEMPLATE_BUCKET))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	t := template.Must(template.New(templatePath).Parse(fs))

	err = t.Execute(&body, mailData)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return body.String(), nil
}

// TODO refactor templates to come from google cloud file store
var mailTemplateLocations = map[MailType]string{
	MailConfirmation: "confirm_mail.html",
	PassReset:        "password_reset.html",
}
