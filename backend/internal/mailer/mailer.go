package mailer

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
)

// Mailer sends transactional emails.
type Mailer interface {
	Send(to, subject, htmlBody string) error
}

// SMTPMailer sends via SMTP. Works with SendGrid, Resend, Mailgun, and plain SMTP.
type SMTPMailer struct {
	Host string
	Port string
	From string
	User string
	Pass string
}

func (m *SMTPMailer) Send(to, subject, htmlBody string) error {
	addr := m.Host + ":" + m.Port
	auth := smtp.PlainAuth("", m.User, m.Pass, m.Host)

	fromAddress := m.From
	if parsed, err := mail.ParseAddress(m.From); err == nil {
		fromAddress = parsed.Address
	}

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", m.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	return smtp.SendMail(addr, auth, fromAddress, []string{to}, []byte(msg))
}

// SentMail is one recorded email.
type SentMail struct {
	To      string
	Subject string
	Body    string
}

// MockMailer records sent mail for use in tests.
type MockMailer struct {
	mu   sync.Mutex
	Sent []SentMail
}

func (m *MockMailer) Send(to, subject, htmlBody string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sent = append(m.Sent, SentMail{To: to, Subject: subject, Body: htmlBody})
	return nil
}
