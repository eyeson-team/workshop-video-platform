package goose

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"time"
)

type Login struct {
	ID        uint
	AuthKey   string
	ExpiresAt time.Time
	UserID    uint
	User      User
}

// SendMail delivers an email with the authentication link to the given users
// email address.
func (l *Login) SendMail(baseURL string) error {
	body := fmt.Sprintf("Your session Login:\r\n  %v/sessions/%v\r\n", baseURL, l.AuthKey)
	var user User
	db.Model(&l).Association("User").Find(&user)
	return sendMail(user.Email, "MeetDuck - Your Login Link", body)
}

func sendMail(to, subject, body string) error {
	log.Printf("Sending mail to %s\n", to)

	from := "meetduck@eyeson.com"
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n\r\n"+
		"%s\r\n", from, to, subject, body))

	password := os.Getenv("MAIL_SMTP_PWD")
	if len(password) == 0 {
		return sendLocalMail(to, from, msg)
	}

	smtpHost := "smtp.googlemail.com"
	smtpPort := 587
	return smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort),
		smtp.PlainAuth("", from, password, smtpHost), from, []string{to}, msg)
}

func sendLocalMail(to, from string, msg []byte) error {
	return smtp.SendMail("localhost:1025", nil, from, []string{to}, msg)
}

func NewLogin(user *User) (*Login, error) {
	var login Login
	key, err := generateSecureURLSafeKey(32)
	if err != nil {
		return nil, err
	}
	login.UserID = user.ID
	login.AuthKey = key
	login.ExpiresAt = time.Now().Add(time.Minute * 15)
	result := db.Create(&login)
	return &login, result.Error
}

// generateSecureURLSafeKey generates a secure random key of given length that
// is URL safe.
func generateSecureURLSafeKey(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
