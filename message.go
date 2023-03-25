package mail

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type Message struct {
	Options MailOptions
	From    string
	To      []string
	Subject string
	Body    []byte
	Cc      []string
	Bcc     []string
}

var ErrEmptyFrom = errors.New("from is empty")
var ErrEmptyTo = errors.New("to is empty")
var ErrEmptySubject = errors.New("subject is empty")
var ErrEmptyBody = errors.New("body is empty")

// Send email to list of recipient with subject and body message
func (msg *Message) Send() error {
	if msg.From == "" {
		return fmt.Errorf("message: %w", ErrEmptyFrom)
	}

	if len(msg.To) == 0 {
		return fmt.Errorf("message: %w", ErrEmptyTo)
	}

	if msg.Subject == "" {
		return fmt.Errorf("message: %w", ErrEmptySubject)
	}

	if len(msg.Body) == 0 {
		return fmt.Errorf("message: %w", ErrEmptyBody)
	}

	var msgBuilder strings.Builder

	msgBuilder.WriteString("From: " + msg.From + "\r\n")
	msgBuilder.WriteString("To: " + strings.Join(msg.To, ",") + "\r\n")
	msgBuilder.WriteString("Subject: " + msg.Subject + "\r\n")
	msgBuilder.WriteString("Message-ID: " + generateMessageID() + "\r\n")
	msgBuilder.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")

	if len(msg.Cc) > 0 {
		msgBuilder.WriteString("Cc: " + strings.Join(msg.Cc, ",") + "\r\n")
	}

	if len(msg.Bcc) > 0 {
		msgBuilder.WriteString("Bcc: " + strings.Join(msg.Bcc, ",") + "\r\n")
	}

	if len(msg.Body) > 0 {
		msgBuilder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		msgBuilder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		msgBuilder.WriteString("\r\n")
		msgBuilder.Write(msg.Body)
	}

	auth, err := msg.Options.plainAuth()
	if err != nil {
		return err
	}

	err = smtp.SendMail(msg.Options.Host+":"+msg.Options.Port, auth, msg.From, msg.To, []byte(msgBuilder.String()))
	if err != nil {
		return err
	}

	return nil
}
