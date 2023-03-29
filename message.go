package mail

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Attachment struct {
	Name        string
	Data        []byte
	ContentType string
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

func (m *Message) SetFrom(from string) {
	m.from = from
}

func (m *Message) SetTo(to []string) {
	m.to = to
}

func (m *Message) SetSubject(subject string) {
	m.subject = subject
}

func (m *Message) SetCc(cc []string) {
	m.cc = cc
}

func (m *Message) SetBcc(bcc []string) {
	m.bcc = bcc
}

func (m *Message) SetBodyPlainText(content []byte) {
	m.plainText = content
}

func (m *Message) SetBodyHTML(content []byte) {
	m.html = content
}

// SetAttachment set attachment to email message
// param name is either path to file or url
func (m *Message) SetAttachment(filename string) error {
	if filename == "" {
		return fmt.Errorf("message: %w", ErrEmptyAttachment)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("message: %w", ErrFileNotFound)
	}

	defer file.Close()

	contentType := filepath.Ext(filename)

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	fileBuffer := make([]byte, fileSize)

	file.Read(fileBuffer)

	attachment := Attachment{
		Name:        fileInfo.Name(),
		Data:        fileBuffer,
		ContentType: contentType,
	}

	m.attachment = append(m.attachment, attachment)

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

// func (m *Message) SendMultipart() error {
// 	if m.from == "" {
// 		return fmt.Errorf("message: %w", ErrEmptyFrom)
// 	}

// 	if len(m.to) == 0 {
// 		return fmt.Errorf("message: %w", ErrEmptyTo)
// 	}

// 	if m.subject == "" {
// 		return fmt.Errorf("message: %w", ErrEmptySubject)
// 	}

// 	from, err := mail.ParseAddress(m.from)
// 	if err != nil {
// 		return err
// 	}

// 	for _, to := range m.to {
// 		_, err := mail.ParseAddress(to)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	m.header = "From: " + from.String() + "\r\n" +
// 		"To: " + strings.Join(m.to, ",") + "\r\n" +
// 		"Subject: " + m.subject + "\r\n" +
// 		"Message-ID: " + generateMessageID() + "\r\n" +
// 		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
// 		"MIME-Version: 1.0\r\n"

// 	if len(m.cc) > 0 {
// 		m.header += "Cc: " + strings.Join(m.cc, ",") + "\r\n"
// 	}

// 	if len(m.bcc) > 0 {
// 		m.header += "Bcc: " + strings.Join(m.bcc, ",") + "\r\n"
// 	}

// 	if len(m.attachment) > 0 {
// 		m.header += "Content-Type: multipart/mixed; boundary=\"" + m.boundary + "\""
// 	}

// 	auth, err := m.Options.plainAuth()
// 	if err != nil {
// 		return err
// 	}

// 	log.Println(m.body.String())

// 	body := m.header + m.body.String()
// 	err = smtp.SendMail(m.Options.Host+":"+m.Options.Port, auth, m.from, m.to, []byte(body))
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (m *Message) Send() (err error) {
	if m.from == "" {
		return fmt.Errorf("message: %w", ErrEmptyFrom)
	}

	if len(m.to) == 0 {
		return fmt.Errorf("message: %w", ErrEmptyTo)
	}

	if m.subject == "" {
		return fmt.Errorf("message: %w", ErrEmptySubject)
	}

	from, err := mail.ParseAddress(m.from)
	if err != nil {
		return err
	}

	for _, to := range m.to {
		_, err := mail.ParseAddress(to)
		if err != nil {
			return err
		}
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	var headContentType string

	if len(m.plainText) > 0 {
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": {"text/plain; charset=UTF-8"},
		})
		if err != nil {
			return err
		}

		_, err = part.Write(m.plainText)
		if err != nil {
			return err
		}

		headContentType = "text/plain; charset=UTF-8"
	}

	if len(m.html) > 0 {
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": {"text/html; charset=UTF-8"},
		})
		if err != nil {
			return err
		}

		_, err = part.Write(m.html)
		if err != nil {
			return err
		}

		headContentType = "text/html; charset=UTF-8"

		if len(m.plainText) != 0 {
			headContentType = "multipart/alternative"
		}
	}

	if len(m.attachment) > 0 {
		for _, attachment := range m.attachment {
			part, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Type":        {attachment.ContentType},
				"Content-Disposition": {fmt.Sprintf("attachment; filename=\"%s\"", attachment.Name)},
			})
			if err != nil {
				return err
			}

			_, err = part.Write(attachment.Data)
			if err != nil {
				return err
			}
		}

		headContentType = "multipart/mixed; boundary=\"" + writer.Boundary() + "\""
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	m.header = "From: " + from.String() + "\r\n" +
		"To: " + strings.Join(m.to, ",") + "\r\n" +
		"Subject: " + m.subject + "\r\n" +
		"Message-ID: " + generateMessageID() + "\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: " + headContentType + "\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n"

	if len(m.cc) > 0 {
		m.header += "Cc: " + strings.Join(m.cc, ",") + "\r\n"
	}

	if len(m.bcc) > 0 {
		m.header += "Bcc: " + strings.Join(m.bcc, ",") + "\r\n"
	}

	message := m.header + "\r\n" + body.String()

	fmt.Println(message)

	// auth, err := m.Options.plainAuth()
	// if err != nil {
	// 	return err
	// }

	// err = smtp.SendMail(m.Options.Host+":"+m.Options.Port, auth, m.from, m.to, []byte(message))
	// if err != nil {
	// 	return err
	// }

	return nil
}
