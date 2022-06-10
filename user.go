package goose

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type User struct {
	ID         uint
	Name       string
	Email      string
	Workspaces []*Workspace `gorm:"many2many:workspace_users;"`
}

// Assign does create an association bettwen given user and workspace.
func (u *User) Assign(workspace *Workspace) {
	db.Model(&u).Association("Workspaces").Append(workspace)
	db.Save(&u)
}

// Allowed does flag if a user has access to a given workspace.
func (u *User) Allowed(w *Workspace) bool {
	return db.Model(&u).Where("id = ?", w.ID).Association("Workspaces").Count() == 1
}

// func (u *User) FindWorkspace(id int) (*Workspace, error) {
// }

// NewUser fetches or creates a user by a given email address. Note that this
// method does handle persistence.
func NewUser(email string) (*User, error) {
	user := User{Email: email, Name: ExtractNameFromEmail(email)}
	result := db.Where(&User{Email: email}).FirstOrCreate(&user)
	return &user, result.Error
}

// ExtractNameFromEmail does reformat the first part of an email to guess a
// name out of it, such that `freddie.krueger@example.com` becomes `Freddie
// Krueger`.
func ExtractNameFromEmail(email string) string {
	name := strings.Split(email, "@")[0]
	name = strings.ReplaceAll(name, ".", " ")
	c := cases.Title(language.English)
	return c.String(name)
}
