package api_test

type EmailRecord struct {
	To      string
	Subject string
	Html    string
}

type MockEmailSender struct {
	Sent      []EmailRecord
	SendError error
}

func (m *MockEmailSender) Send(to, subject, html string) error {
	m.Sent = append(m.Sent, EmailRecord{
		To:      to,
		Subject: subject,
		Html:    html,
	})
	return m.SendError
}
