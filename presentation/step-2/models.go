package goose

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"time"
)

type Workspace struct {
	ID    uint
	Topic string
}

type User struct {
	ID    uint
	Name  string
	Email string
}

func NewUser(email string) (*User, error) {
	// ...
}

type Login struct {
	ID        int
	AuthCode  string
	ExpiresAt time.Time
	UserID    uint
	User      User
}

func (l *Login) SendMail(baseURL string) error {
	var user User
	db.Model(&l).Association("User").Find(&user)

	// https://localhost:8077/login/:authcode
	body := fmt.Sprintf("Your session Login:\r\n %v/login/%v\r\n", baseURL, l.AuthCode)
	return sendMail(user.Email, "MeetDuck - Your Login Link", body)
}

func sendMail(to, subject, body string) error {
	log.Printf("Send Email to %v with %v", to, body)

	from := "meetduck@eyeson.com"
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n\r\n"+
		"%s\r\n", from, to, subject, body))
	return smtp.SendMail("127.0.0.1:1025", nil, from, []string{to}, msg)
}

func NewLogin(user *User) (*Login, error) {
	var login Login
	code, err := generateSecureURLSafeKey(32)
	if err != nil {
		return nil, err
	}
	login.UserID = user.ID
	login.AuthCode = code
	login.ExpiresAt = time.Now().Add(time.Minute * 15)
	result := db.Create(&login)
	return &login, result.Error
}

func generateSecureURLSafeKey(n uint) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
