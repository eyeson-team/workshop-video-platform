package goose

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Workspace struct {
	ID         uint
	Topic      string
	Meetings   []Meeting
	Recordings []Recording
}

func (w *Workspace) NewMeeting(startedAt time.Time) (*Meeting, error) {
	meeting := Meeting{Workspace: *w, StartedAt: startedAt}
	qry := db.Where("started_at = ? AND workspace_id = ?", startedAt, w.ID)
	result := qry.FirstOrCreate(&meeting)
	return &meeting, result.Error
}

func (w *Workspace) LastMeeting() (*Meeting, error) {
	var meeting Meeting
	result := db.Where("workspace_id = ?", w.ID).Last(&meeting)
	return &meeting, result.Error
}

func FindWorkspace(id string) (*Workspace, error) {
	var workspace Workspace
	result := db.First(&workspace, "id = ?", id)
	return &workspace, result.Error
}

type User struct {
	ID    uint
	Email string
	Name  string
}

// NewUser creates or finds an user by given email address.
func NewUser(email string) (*User, error) {
	if strings.HasSuffix(email, "@eyeson.com") == false {
		return nil, errors.New("Emails are restricted to company domain")
	}
	user := User{Email: email, Name: ExtractNameFromEmail(email)}
	result := db.Where(&User{Email: email}).FirstOrCreate(&user)
	return &user, result.Error
}

// christoph.lipautz@eyeson.com => Christoph Lipautz
func ExtractNameFromEmail(email string) string {
	name := strings.Split(email, "@")[0]
	name = strings.ReplaceAll(name, ".", " ")
	c := cases.Title(language.English)
	return c.String(name)
}

type Login struct {
	ID        uint
	AuthKey   string
	ExpiresAt time.Time
	UserID    uint
	User      User
}

func (l *Login) SendMail(baseURL string) error {
	var user User
	db.Model(&l).Association("User").Find(&user)
	body := fmt.Sprintf("You session login:\r\n %v/login/%v\r\n", baseURL, l.AuthKey)
	return sendMail(user.Email, "MeetDuck - Your Login Link", body)
}

func sendMail(to, subject, body string) error {
	// log.Println("[Email Delivery]", to, body)
	from := "meetduck@eyeson.com"
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n\r\n"+
		"%s\r\n", from, to, subject, body))

	return smtp.SendMail("127.0.0.1:1025", nil, from, []string{to}, msg)
}

func NewLogin(user *User) (*Login, error) {
	key, err := GenerateSecureURLSafeKey(32)
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(time.Minute * 15)
	login := Login{AuthKey: key, ExpiresAt: expires, UserID: user.ID}
	result := db.Create(&login)
	return &login, result.Error
}

func GenerateSecureURLSafeKey(n uint) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type Meeting struct {
	ID          uint
	Reference   string
	StartedAt   time.Time
	EndedAt     time.Time
	WorkspaceID uint
	Workspace   Workspace
}

func (m *Meeting) IsActive() bool {
	return m.EndedAt.IsZero()
}

func (m *Meeting) StoreShutdown() error {
	m.EndedAt = time.Now().UTC()
	return db.Save(&m).Error
}

type Recording struct {
	ID          uint
	Reference   string
	Duration    uint
	StartedAt   time.Time
	EndedAt     time.Time
	WorkspaceID uint
	Workspace   Workspace
}

func (r *Recording) StoragePath() string {
	baseDir := "/tmp/meetduck"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0755)
	}

	return fmt.Sprintf("%s/%d-%d.webm", baseDir, r.WorkspaceID, r.ID)
}

func (r *Recording) Store(downloadLink string) error {
	dest, err := os.Create(r.StoragePath())
	if err != nil {
		return err
	}
	defer dest.Close()

	resp, err := http.Get(downloadLink)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download. Status %d", resp.StatusCode)
	}

	_, err = io.Copy(dest, resp.Body)
	return err
}
