package email

/*
	wrapper for the SendGrid email API
*/

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EMailer struct {
	key         string
	fromEmail   string
	fromUser    string
	attachments []*mail.Attachment
}

func NewEMailer(sgkey, fromEmail, fromUser string) *EMailer {
	var this EMailer
	this.key = sgkey
	this.fromEmail = fromEmail
	this.fromUser = fromUser
	return &this
}

// AddAttachment attach a file to the email  (ex "text/plain")
func (e *EMailer) AddAttachment(data []byte, filename string, filetype string) {
	attachment := mail.NewAttachment()
	encoded := base64.StdEncoding.EncodeToString(data)
	attachment.SetContent(encoded)
	attachment.SetType(filetype)
	attachment.SetFilename(filename)
	attachment.SetDisposition("attachment")
	e.attachments = append(e.attachments, attachment)
}

func (e *EMailer) SendTextMsg(subject, dest, content string) error {
	m := CreatePlainMessage(subject, dest, e.fromUser, e.fromEmail, content)
	m.Attachments = e.attachments
	return Send(m, e.key)
}

func (e *EMailer) SendHTMLMsg(subject, dest, content string) error {
	m := CreateHTMLMessage(subject, dest, e.fromUser, e.fromEmail, content)
	m.Attachments = e.attachments
	return Send(m, e.key)
}

// CreatePlainMessage - creates initial email object using plain text
// Note that dest parameter accepts a comma delimited list of email addresses
func CreatePlainMessage(subject, dest, fromUser, fromEmail, content string) *mail.SGMailV3 {
	m := createMessage(subject, dest, fromUser, fromEmail)
	m.AddContent(mail.NewContent("text/plain", content))
	return m
}

// CreateHTMLMessage - creates initial email object using HTML text
// Note that dest parameter accepts a comma delimited list of email addresses
func CreateHTMLMessage(subject, dest, fromUser, fromEmail, content string) *mail.SGMailV3 {
	m := createMessage(subject, dest, fromUser, fromEmail)
	m.AddContent(mail.NewContent("text/html", content))
	return m
}

func createMessage(subject, dest, fromUser, fromEmail string) *mail.SGMailV3 {
	m := new(mail.SGMailV3)
	m.SetFrom(mail.NewEmail(fromUser, fromEmail))
	m.Subject = subject
	p := mail.NewPersonalization()

	a := strings.Split(dest, ",")
	for _, to := range a {
		p.AddTos(mail.NewEmail("", to))
	}
	m.AddPersonalizations(p)
	return m
}

// Send - launches the prepared email, using the provided sendgrid api key
func Send(m *mail.SGMailV3, key string) error {
	client := sendgrid.NewSendClient(key)
	response, err := client.Send(m)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusAccepted {
		return errors.New(fmt.Sprintf(`Send email failed with status code %d`, response.StatusCode))
	}
	return nil
}
