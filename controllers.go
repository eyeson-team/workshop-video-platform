package goose

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	eyeson "github.com/eyeson-team/eyeson-go"
	"github.com/gofiber/fiber/v2"
)

// authenticated wraps handler functions, redirects if no user can be fetched
// from context cookie or forwards the request context and user.
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

// readUserFromCookie provides the current user to be fetched from cookie and
// database. If an error occurs during the reading of the user, all cookies
// will be cleared. This prevents issues when a user gets deleted, or the
// cookie has become invalid.
func readUserFromCookie(c *fiber.Ctx) (*User, error) {
	id := c.Cookies("user", "0")
	if id == "0" {
		return nil, errors.New("No cookie present")
	}
	var user User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		c.ClearCookie()
		return nil, err
	}
	return &user, nil
}

// HandleRootPath decides on the presence of a user session to either show the
// workspaces of the user or render a login form.
func HandleRootPath(c *fiber.Ctx) error {
	user, err := readUserFromCookie(c)
	if err != nil {
		return LoginsNew(c)
	}
	return WorkspacesIndex(c, user)
}

// LoginsNew does render the login form and greets the visitor with a random
// welcome message.
func LoginsNew(c *fiber.Ctx) error {
	messages := []string{
		"Welcome to MeetDuck and have a nice day!",
		"Join some ProDucktive Meetings!",
	}
	message := messages[rand.Intn(len(messages))]
	return c.Render("sessions/login", fiber.Map{"Message": message})
}

// LoginsCreate does generate an authentication link and sends it to the users
// email address.
func LoginsCreate(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if strings.HasSuffix(email, "@"+AUTHORIZED_DOMAIN) == false {
		log.Printf("Unauthorized login attempt with email %v\n", email)
		return fiber.ErrForbidden
	}

	user, err := NewUser(email)
	if err != nil {
		log.Println("Failed to create user:", err)
		return fiber.ErrBadRequest
	}
	login, err := NewLogin(user)
	if err != nil {
		log.Println("Failed to create login:", err)
		return fiber.ErrBadRequest
	}
	if err := login.SendMail(c.BaseURL()); err != nil {
		log.Println("Failed to send mail:", err)
		return fiber.ErrInternalServerError
	}
	return c.Render("sessions/instructions", nil)
}

// LoginsShow does take a authentication key, validates it and creates a new
// session for the given user on success.
func LoginsShow(c *fiber.Ctx) error {
	var login Login
	result := db.First(&login, "auth_key = ? AND expires_at >= ?", c.Params("auth"), time.Now())
	if result.Error != nil {
		log.Println(result.Error)
		return fiber.ErrBadRequest
	}
	log.Printf("Login of user %d\n", login.UserID)
	c.Cookie(&fiber.Cookie{Name: "user", Value: fmt.Sprintf("%d", login.UserID)})
	db.Delete(&login)
	return c.Redirect("/", 303)
}

// WorkspacesIndex renders a HTML site with all workspaces registered.
func WorkspacesIndex(c *fiber.Ctx, user *User) error {
	workspaces := []Workspace{}
	db.Model(&user).Association("Workspaces").Find(&workspaces)
	return c.Render("workspaces/index", fiber.Map{
		"User":       user,
		"Workspaces": workspaces,
	})
}

// WorkspacesNew renders the form to create a new workspace.
func WorkspacesNew(c *fiber.Ctx, user *User) error {
	return c.Render("workspaces/new", fiber.Map{"User": user})
}

// WorkspacesCreate creates a new workspace with the given form values.
//
// NOTE: We could stick to validator package instead of validating the input
// length.
func WorkspacesCreate(c *fiber.Ctx, user *User) error {
	topic := c.FormValue("topic")
	if len(topic) < 3 || len(topic) > 255 {
		return fiber.ErrBadRequest
	}
	content := c.FormValue("content")
	workspace := Workspace{Topic: topic, Content: content}
	db.Create(&workspace)
	user.Assign(&workspace)
	db.Save(&user)

	return c.Redirect(fmt.Sprintf("/workspaces/%d", workspace.ID), 303)
}

// WorkspacesEdit renders the form to edit an existing workspace.
func WorkspacesEdit(c *fiber.Ctx, user *User) error {
	var workspace Workspace
	err := db.First(&workspace, c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if user.Allowed(&workspace) == false {
		return fiber.ErrForbidden
	}
	var users []User
	db.Where("id != ?", user.ID).Find(&users)
	return c.Render("workspaces/edit", fiber.Map{
		"User":      user,
		"Users":     users,
		"Workspace": workspace,
	})
}

// WorkspacesAddUser assigns a new user to a workspace.
func WorkspacesAddUser(c *fiber.Ctx, user *User) error {
	id := c.Params("id")
	var workspace Workspace
	err := db.First(&workspace, id)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if user.Allowed(&workspace) == false {
		return fiber.ErrForbidden
	}
	var colleague User
	db.Find(&colleague, c.FormValue("user_id"))
	colleague.Assign(&workspace)
	return c.Redirect(fmt.Sprintf("/workspaces/%v", id), 303)
}

// WorkspacesUpdate updates an existing workspace.
func WorkspacesUpdate(c *fiber.Ctx, user *User) error {
	var workspace Workspace
	err := db.First(&workspace, c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if user.Allowed(&workspace) == false {
		return fiber.ErrForbidden
	}
	topic := c.FormValue("topic")
	if len(topic) < 3 || len(topic) > 255 {
		return fiber.ErrBadRequest
	}
	workspace.Topic = topic
	workspace.Content = c.FormValue("content")
	db.Save(&workspace)

	return c.Render("workspaces/show", fiber.Map{
		"User":      user,
		"Workspace": workspace,
	})
}

// WorkspacesShow renders a single workspaces and its assets (meetings,
// recordings, snapshots).
func WorkspacesShow(c *fiber.Ctx, user *User) error {
	var workspace Workspace
	result := db.Model(&workspace).Preload("Meetings").Preload("Users").
		Preload("Recordings").Preload("Snapshots").Find(&workspace, c.Params("id"))

	if result.Error != nil {
		log.Println(result.Error)
		return fiber.ErrBadRequest
	}
	if user.Allowed(&workspace) == false {
		return fiber.ErrForbidden
	}

	return c.Render("workspaces/show", fiber.Map{
		"User":      user,
		"Workspace": workspace,
	})
}

// MeetingsEdit provides a form to edit a meeting.
func MeetingsEdit(c *fiber.Ctx, user *User) error {
	var meeting Meeting
	err := db.Preload("Workspace").Find(&meeting, c.Params("id")).Error
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	if user.Allowed(&meeting.Workspace) == false {
		return fiber.ErrForbidden
	}
	return c.Render("meetings/edit", fiber.Map{
		"User":    user,
		"Meeting": meeting,
	})
}

// MeetingsUpdate updates an existing meeting and redirects user back to
// the workspace.
func MeetingsUpdate(c *fiber.Ctx, user *User) error {
	var meeting Meeting
	err := db.Preload("Workspace").First(&meeting, c.Params("id")).Error
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	if user.Allowed(&meeting.Workspace) == false {
		return fiber.ErrForbidden
	}
	meeting.Content = c.FormValue("content")
	db.Save(&meeting)

	return c.Redirect(fmt.Sprintf("/workspaces/%d", meeting.WorkspaceID), 303)
}

// MeetingsCreate does start a meeting at the eyeson service API and either
// forward the user to the eyeson default web UI or render the custom web UI.
// The eyeson service receives two options:
//  - sfu_mode=disabled enforces the single stream that make the custom UI
//    easier to build.
//  - exit_url=<absolute-url-to-workspace> does allow the eyeson default web UI
//    to redirect users back to our platform when they leave the meeting.
func MeetingsCreate(c *fiber.Ctx, user *User) error {
	id := c.Params("id")

	var workspace Workspace
	result := db.First(&workspace, id)
	if result.Error != nil {
		return fiber.ErrBadRequest
	}
	if user.Allowed(&workspace) == false {
		return fiber.ErrForbidden
	}
	options := map[string]string{
		"options[sfu_mode]": "disabled",
		"options[exit_url]": c.BaseURL() + "/workspaces/" + id,
	}
	room, err := videoService.Rooms.Join(id, "user", options)
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}

	if string(c.FormValue("mode")) == "forward" {
		return c.Redirect(room.Data.Links.Gui, 303)
	}
	return c.Render("meetings/show", fiber.Map{
		"User":      user,
		"Workspace": workspace,
		"AccessKey": room.Data.AccessKey,
	})
}

// WebhooksCreate reads a webhook and stores a new event. On error we still
// respond with a successfull HTTP status such that the webhook does know it
// correctly delivered the data to our endpoint.
//
// The signature will ensure the webhook has been sent by eyeson.
//
// We do store and associate webhook details for meeting events and recordings.
func WebhooksCreate(c *fiber.Ctx) error {
	body := c.Body()

	signature := c.GetReqHeaders()["X-Eyeson-Signature"]
	if err := validateWebhook(body, signature); err != nil {
		log.Println(err)
		return nil
	}

	var webhook eyeson.Webhook
	if err := json.Unmarshal(c.Body(), &webhook); err != nil {
		log.Println("Could not parse webhook request.")
		return nil
	}
	log.Printf("Store new webhook %s for room %s\n", webhook.Type, webhook.Room.Id)
	if err := storeWebhookEvent(&webhook); err != nil {
		log.Println("Failed to store webhook,", err)
	}
	return nil
}

// RecordingsShow streams a recording video file.
func RecordingsShow(c *fiber.Ctx, user *User) error {
	var recording Recording
	result := db.Model(&recording).Preload("Workspace").
		First(&recording, c.Params("id"))

	if result.Error != nil {
		log.Println(result.Error)
		return fiber.ErrBadRequest
	}
	if user.Allowed(&recording.Workspace) == false {
		return fiber.ErrForbidden
	}

	return c.SendFile(recording.StoragePath())
}

// SnapshotsShow renders a single snapshot image.
func SnapshotsShow(c *fiber.Ctx, user *User) error {
	var snapshot Snapshot
	result := db.Model(&snapshot).Preload("Workspace").
		Find(&snapshot, c.Params("id"))

	if result.Error != nil {
		log.Println(result.Error)
		return fiber.ErrBadRequest
	}
	if user.Allowed(&snapshot.Workspace) == false {
		return fiber.ErrForbidden
	}

	return c.SendFile(snapshot.StoragePath())
}
