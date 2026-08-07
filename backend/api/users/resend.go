package users

import "github.com/resend/resend-go/v3"

type EmailSender interface {
	Send(to, subject, html string) error
}
type ResendSender struct {
	Client *resend.Client
	From   string
}

func (rs *ResendSender) Send(to, subject, html string) error {
	params := &resend.SendEmailRequest{
		From:    rs.From,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	_, err := rs.Client.Emails.Send(params)
	return err
}
