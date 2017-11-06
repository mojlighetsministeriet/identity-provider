package email

import (
	"crypto/tls"

	gomail "gopkg.in/gomail.v2"
)

// SMTPSender is used to send emails
type SMTPSender struct {
	Host      string
	Port      int
	Email     string
	Password  string
	TLSConfig *tls.Config
}

// Send will send an email
func (sender *SMTPSender) Send(to string, subject string, body string) (err error) {
	if sender.Port == 0 {
		sender.Port = 587
	}

	message := gomail.NewMessage()
	message.SetHeader("From", sender.Email)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	dialer := gomail.NewDialer(sender.Host, sender.Port, sender.Email, sender.Password)
	dialer.TLSConfig = sender.TLSConfig
	err = dialer.DialAndSend(message)

	return
}
