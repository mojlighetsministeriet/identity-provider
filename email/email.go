package email

import (
	"bytes"
	"crypto/tls"
	"errors"
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

// Templates can hold templates
type Templates struct {
	templates []*Template
	Sender    *SMTPSender
}

// Add adds a new template
func (templates *Templates) Add(emailTemplate Template) {
	emailTemplate.Compile()
	templates.templates = append(templates.templates, &emailTemplate)
}

// Render will generate the subject and body strings from a template
func (templates *Templates) Render(name string, subjectData interface{}, bodyData interface{}) (subject string, body string, err error) {
	var templateToRender *Template

	for _, emailTemplate := range templates.templates {
		if emailTemplate.Name == name {
			templateToRender = emailTemplate
			break
		}
	}

	if templateToRender == nil {
		err = errors.New("Template " + name + " is not registered")
		return
	}

	subject, err = templateToRender.GetSubject(subjectData)
	if err != nil {
		subject = ""
		return
	}

	body, err = templateToRender.GetBody(bodyData)
	if err != nil {
		subject = ""
		body = ""
	}

	return
}

// RenderAndSend will render a template and send an email
func (templates *Templates) RenderAndSend(to string, name string, subjectData interface{}, bodyData interface{}) (err error) {
	subject, body, err := templates.Render(name, subjectData, bodyData)
	if err != nil {
		return
	}

	err = templates.Sender.Send(to, subject, body)
	return
}

// Template lets you format emails with go templates
type Template struct {
	Name            string
	Subject         string
	subjectTemplate *template.Template
	Body            string
	bodyTemplate    *template.Template
}

// Compile will re-compile the template Subject and Body to templates
func (emailTemplate *Template) Compile() (err error) {
	emailTemplate.subjectTemplate, err = template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return
	}

	emailTemplate.bodyTemplate, err = template.New("body").Parse(emailTemplate.Body)
	return
}

// GetSubject will return the populated Subject by passing a data map to the function
func (emailTemplate *Template) GetSubject(data interface{}) (output string, err error) {
	buffer := new(bytes.Buffer)

	if emailTemplate.subjectTemplate == nil {
		emailTemplate.Compile()
	}

	err = emailTemplate.subjectTemplate.Execute(buffer, data)
	if err != nil {
		return
	}

	output = buffer.String()

	return
}

// GetBody will return the populated Body by passing a data map to the function
func (emailTemplate *Template) GetBody(data interface{}) (output string, err error) {
	buffer := new(bytes.Buffer)

	if emailTemplate.subjectTemplate == nil {
		emailTemplate.Compile()
	}

	err = emailTemplate.subjectTemplate.Execute(buffer, data)
	if err != nil {
		return
	}

	output = buffer.String()

	return
}
