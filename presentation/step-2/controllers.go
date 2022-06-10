package goose

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func HandleRootPath(c *fiber.Ctx) error {
	user, err := readUserFromCookie(c)
	if err != nil {
		return c.Render("logins/new", fiber.Map{})
	}
	return WorkspacesIndex(c, user)
}

func readUserFromCookie(c *fiber.Ctx) (*User, error) {
	id := c.Cookies("user", "")
	if id == "" {
		return nil, errors.New("Cookie not set")
	}
	var user User
	result := db.First(&user, "id = ?", id)
	return &user, result.Error
}

func authenticated(f func(c *fiber.Ctx, user *User) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user, err := readUserFromCookie(c)
		if err != nil {
			c.ClearCookie()
			return c.Redirect("/", 303)
		}
		return f(c, user)
	}
}

func WorkspacesIndex(c *fiber.Ctx, user *User) error {
	var workspaces []Workspace
	db.Find(&workspaces)

	return c.Render("workspaces/index", fiber.Map{
		"User":       user,
		"Workspaces": workspaces,
	})
}

func WorkspacesShow(c *fiber.Ctx, user *User) error {
	var workspace Workspace
	result := db.First(&workspace, "id = ?", c.Params("id"))
	if result.Error != nil {
		return fiber.ErrBadRequest
	}
	return c.Render("workspaces/show", fiber.Map{
		"User":      user,
		"Workspace": workspace,
	})
}

func WorkspacesCreate(c *fiber.Ctx, user *User) error {
	db.Create(&Workspace{Topic: c.FormValue("topic")})
	return c.Redirect("/", 303)
}

func MeetingsCreate(c *fiber.Ctx, user *User) error {
	id := c.Params("id")
	var workspace Workspace
	if err := db.First(&workspace, "id = ?", id).Error; err != nil {
		return fiber.ErrNotFound
	}
	options := map[string]string{
		"name":              workspace.Topic,
		"options[sfu_mode]": "disabled",
		"options[exit_url]": c.BaseURL() + "/",
	}
	meeting, err := videoService.Rooms.Join(id, user.Name, options)
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	return c.Redirect(meeting.Data.Links.Gui, 303)
}

func LoginsCreate(c *fiber.Ctx) error {
	email := c.FormValue("email")
	// SKIPPED: validate email
	if strings.HasSuffix(email, "@eyeson.com") == false {
		return fiber.ErrForbidden
	}
	user, err := NewUser(email)
	if err != nil {
		return fiber.ErrBadRequest
	}
	login, err := NewLogin(user)
	if err != nil {
		return fiber.ErrBadRequest
	}
	login.SendMail(c.BaseURL())
	return c.Render("logins/instructions", fiber.Map{})
}

func LoginsShow(c *fiber.Ctx) error {
	authCode := c.Params("auth")
	var login Login
	err := db.First(&login, "auth_code = ?", authCode).Error
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	c.Cookie(&fiber.Cookie{Name: "user", Value: fmt.Sprintf("%d", login.UserID)})
	db.Delete(&login)
	return c.Redirect("/", 303)
}

func LoginsDelete(c *fiber.Ctx) error {
	c.ClearCookie()
	return c.Redirect("/", 303)
}
