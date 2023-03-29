package mail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/smtp"
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
	Options    Options
	from       string
	to         []string
	subject    string
	cc         []string
	bcc        []string
	header     string
	plainText  []byte
	html       []byte
	attachment []Attachment
}

var ErrEmptyFrom = errors.New("from is empty")
var ErrEmptyTo = errors.New("to is empty")
var ErrEmptySubject = errors.New("subject is empty")
var ErrEmptyBody = errors.New("body is empty")
var ErrEmptyAttachment = errors.New("attachment is empty")
var ErrFileNotFound = errors.New("file not found")

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
	contentType := mime.TypeByExtension(filepath.Ext(filename))

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	fileBuffer := make([]byte, fileSize)

	file.Read(fileBuffer)

	// contentType := http.DetectContentType(fileBuffer)
	attachment := Attachment{
		Name:        fileInfo.Name(),
		Data:        fileBuffer,
		ContentType: contentType,
	}

	m.attachment = append(m.attachment, attachment)

	return nil
}

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
			headContentType = "multipart/alternative; boundary=\"" + writer.Boundary() + "\""
		}
	}

	if len(m.attachment) > 0 {
		for _, attachment := range m.attachment {
			part, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {attachment.ContentType},
				"Content-Transfer-Encoding": {"base64"},
				"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", attachment.Name)},
			})
			if err != nil {
				return err
			}

			encoded := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Data)))
			base64.StdEncoding.Encode(encoded, attachment.Data)

			_, err = part.Write(base64LineBreaker(encoded))
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
		"Content-Type: " + headContentType + "\r\n"

	if len(m.cc) > 0 {
		m.header += "Cc: " + strings.Join(m.cc, ",") + "\r\n"
	}

	if len(m.bcc) > 0 {
		m.header += "Bcc: " + strings.Join(m.bcc, ",") + "\r\n"
	}

	message := m.header + "\r\n" + body.String()

	auth, err := m.Options.plainAuth()
	if err != nil {
		return err
	}

	err = smtp.SendMail(m.Options.Host+":"+m.Options.Port, auth, m.from, m.to, []byte(message))
	if err != nil {
		return err
	}

	return nil
}
