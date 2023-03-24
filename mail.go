package mail

import (
	"crypto/rand"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type mailOptions struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (m *mailOptions) plainAuth() smtp.Auth {
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	return auth
}

// Send email to list of recipient with subject and body message
func (m *mailOptions) sendMail(to []string, from, subject string) error {
	var msgBuilder strings.Builder

	msgBuilder.WriteString("From: " + from + "\r\n")
	msgBuilder.WriteString("To: " + strings.Join(to, ",") + "\r\n")
	msgBuilder.WriteString("Subject: " + subject + "\r\n")
	msgBuilder.WriteString("Message-ID: " + generateMessageID() + "\r\n")
	msgBuilder.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	msgBuilder.WriteString("\r\n\r\n")

	err := smtp.SendMail(m.Host+":"+m.Port, m.plainAuth(), from, to, []byte(msgBuilder.String()))
	if err != nil {
		return err
	}

	return nil
}

// Message-ID in header email is consist of uuid and hostname
// Example: <uuid@hostname>
// This header is important for email server to identify email
// Also this header is used to prevent email from being detected as spam
// See: https://tools.ietf.org/html/rfc5322#section-3.6.4
func generateMessageID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	uuid, err := generateUUID()
	if err != nil {
		fmt.Println("Error when generate uuid: ", err.Error())
	}

	messageId := fmt.Sprintf("<%s@%s>", uuid, hostname)

	return messageId
}

// This func is generate uuid v4 that compliant with RFC 4122
// See: https://www.rfc-editor.org/rfc/rfc4122
func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// version 4 (pseudo-random); see section 4.4
	uuid[6] = (uuid[6] & 0x0f) | 0x40

	// variant bits; see section 4.1.1
	uuid[8] = (uuid[8] & 0xbf) | 0x80

	result := fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	return result, nil
}
