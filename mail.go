package mail

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

type Options struct {
	Host     string
	Port     string
	Username string
	Password string
}

var ErrEmptyHost = errors.New("host is empty")
var ErrEmptyPort = errors.New("port is empty")
var ErrEmptyUsername = errors.New("username is empty")
var ErrEmptyPassword = errors.New("password is empty")

// Easiest way to get authentication for smtp server
// See: https://golang.org/pkg/net/smtp/#PlainAuth
func (m *Options) plainAuth() (smtp.Auth, error) {
	if m.Username == "" {
		return nil, fmt.Errorf("mail options: %w", ErrEmptyUsername)
	}

	if m.Password == "" {
		return nil, fmt.Errorf("mail options: %w", ErrEmptyPassword)
	}

	if m.Host == "" {
		return nil, fmt.Errorf("mail options: %w", ErrEmptyHost)
	}

	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	return auth, nil
}

// New is used to create new instance of Message
func New(opt Options) *Message {
	return &Message{
		Options: opt,
	}
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
