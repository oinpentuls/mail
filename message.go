package mail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
	"time"
)

type Attachment struct {
	Name string
	Data []byte
}

type Message struct {
	Options     MailOptions
	from        string
	to          []string
	subject     string
	cc          []string
	bcc         []string
	header      string
	body        bytes.Buffer
	plainText   []byte
	html        []byte
	boundary    string
	contentType ContentType
	attachment  []Attachment
}

type ContentType string

var (
	ContentTypePlainText ContentType = "text/plain"
	ContentTypeHTML      ContentType = "text/html"
	ContentTypeMultipart ContentType = "multipart/mixed"
)

var ErrEmptyFrom = errors.New("from is empty")
var ErrEmptyTo = errors.New("to is empty")
var ErrEmptySubject = errors.New("subject is empty")
var ErrEmptyBody = errors.New("body is empty")
var ErrEmptyAttachment = errors.New("attachment is empty")
var ErrFileNotFound = errors.New("file not found")

type BodyType string

type BodyMessage struct {
	Type    BodyType
	Content []byte
}

func (m *Message) SetFrom(from string) error {
	if from == "" {
		return fmt.Errorf("mail: %w", ErrEmptyFrom)
	}

	m.from = from
	return nil
}

func (m *Message) SetTo(to []string) error {
	if len(to) == 0 {
		return fmt.Errorf("mail: %w", ErrEmptyTo)
	}

	m.to = to
	return nil
}

func (m *Message) SetSubject(subject string) error {
	if subject == "" {
		return fmt.Errorf("mail: %w", ErrEmptySubject)
	}

	m.subject = subject
	return nil
}

func (m *Message) SetCc(cc []string) {
	m.cc = cc
}

func (m *Message) SetBcc(bcc []string) {
	m.bcc = bcc
}

func (msg *Message) SetBodyPlainText(content string) error {
	if content == "" {
		return fmt.Errorf("message: %w", ErrEmptyBody)
	}

	message := "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	_, err := msg.body.WriteString(message + content)

	if err != nil {
		return err
	}

	if msg.boundary != "" {
		_, err := msg.body.WriteString(msg.boundary)

		if err != nil {
			return err
		}
	}

	return nil
}

func (msg *Message) SetBodyHTML(content string) error {
	if content == "" {
		return fmt.Errorf("message: %w", ErrEmptyBody)
	}

	_, err := msg.body.WriteString(content)

	if err != nil {
		return err
	}

	if msg.boundary != "" {
		msg.body.WriteString(msg.boundary)

		if err != nil {
			return err
		}
	}

	return nil
}

// SetAttachment set attachment to email message
// param name is either path to file or url
func (msg *Message) SetAttachment(filename string) error {
	if filename == "" {
		return fmt.Errorf("message: %w", ErrEmptyAttachment)
	}

	file, err := os.ReadFile(filename)

	if err == nil {
		return fmt.Errorf("attachment: %s", ErrFileNotFound)
	}

	msg.contentType = ContentTypeMultipart

	writer := multipart.NewWriter(&msg.body)

	mime := getMimeType(filename)
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", mime+"; name=\""+filename+"\"")
	header.Set("Content-Transfer-Encoding", "base64")
	header.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	msg.boundary = writer.Boundary()

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
	base64.StdEncoding.Encode(encoded, file)
	part.Write(base64LineBreaker(encoded))

	writer.Close()
	return nil
}

// Send email to list of recipient with subject and body message
// func (msg *Message) Send() error {
// 	if msg.From == "" {
// 		return fmt.Errorf("message: %w", ErrEmptyFrom)
// 	}

// 	if len(msg.To) == 0 {
// 		return fmt.Errorf("message: %w", ErrEmptyTo)
// 	}

// 	if msg.Subject == "" {
// 		return fmt.Errorf("message: %w", ErrEmptySubject)
// 	}

// 	var msgBuilder strings.Builder

// 	msgBuilder.WriteString("From: " + msg.From + "\r\n")
// 	msgBuilder.WriteString("To: " + strings.Join(msg.To, ",") + "\r\n")
// 	msgBuilder.WriteString("Subject: " + msg.Subject + "\r\n")
// 	msgBuilder.WriteString("Message-ID: " + generateMessageID() + "\r\n")
// 	msgBuilder.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
// 	msgBuilder.WriteString("MIME-Version: 1.0\r\n")

// 	if len(msg.Cc) > 0 {
// 		msgBuilder.WriteString("Cc: " + strings.Join(msg.Cc, ",") + "\r\n")
// 	}

// 	if len(msg.Bcc) > 0 {
// 		msgBuilder.WriteString("Bcc: " + strings.Join(msg.Bcc, ",") + "\r\n")
// 	}

// 	if len(msg.body) > 0 {
// 		msgBuilder.Write(msg.body)
// 	}

// 	auth, err := msg.Options.plainAuth()
// 	if err != nil {
// 		return err
// 	}

// 	err = smtp.SendMail(msg.Options.Host+":"+msg.Options.Port, auth, msg.From, msg.To, []byte(msgBuilder.String()))
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (msg *Message) SendMultipart() error {
	if msg.from == "" {
		return fmt.Errorf("message: %w", ErrEmptyFrom)
	}

	if len(msg.to) == 0 {
		return fmt.Errorf("message: %w", ErrEmptyTo)
	}

	if msg.subject == "" {
		return fmt.Errorf("message: %w", ErrEmptySubject)
	}

	from, err := mail.ParseAddress(msg.from)
	if err != nil {
		return err
	}

	for _, to := range msg.to {
		_, err := mail.ParseAddress(to)
		if err != nil {
			return err
		}
	}

	msg.header = "From: " + from.String() + "\r\n" +
		"To: " + strings.Join(msg.to, ",") + "\r\n" +
		"Subject: " + msg.subject + "\r\n" +
		"Message-ID: " + generateMessageID() + "\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n"

	if len(msg.cc) > 0 {
		msg.header += "Cc: " + strings.Join(msg.cc, ",") + "\r\n"
	}

	if len(msg.bcc) > 0 {
		msg.header += "Bcc: " + strings.Join(msg.bcc, ",") + "\r\n"
	}

	if len(msg.attachment) > 0 {
		msg.header += "Content-Type: multipart/mixed; boundary=\"" + msg.boundary + "\""
	}

	auth, err := msg.Options.plainAuth()
	if err != nil {
		return err
	}

	log.Println(msg.body.String())

	body := msg.header + msg.body.String()
	err = smtp.SendMail(msg.Options.Host+":"+msg.Options.Port, auth, msg.from, msg.to, []byte(body))
	if err != nil {
		return err
	}

	return nil
}
