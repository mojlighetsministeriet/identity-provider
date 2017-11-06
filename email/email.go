package email

import (
	"bytes"
	"crypto/tls"
	"html/template"

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

type EmailTemplate struct {
	Name            string
	Subject         string
	subjectTemplate *template.Template
	Body            string
	bodyTemplate    *template.Template
}

func (emailTemplate *EmailTemplate) Compile() (err error) {
	emailTemplate.subjectTemplate, err = template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return
	}

	emailTemplate.bodyTemplate, err = template.New("body").Parse(emailTemplate.Body)
	return
}

func (emailTemplate *EmailTemplate) GetSubject(data interface{}) (output string, err error) {
	buffer := new(bytes.Buffer)

	err = emailTemplate.subjectTemplate.Execute(buffer, data)
	if err != nil {
		return
	}

	output = buffer.String()

	return
}

func (emailTemplate *EmailTemplate) GetBody(data interface{}) (output string, err error) {
	buffer := new(bytes.Buffer)

	err = emailTemplate.subjectTemplate.Execute(buffer, data)
	if err != nil {
		return
	}

	output = buffer.String()

	return
}
