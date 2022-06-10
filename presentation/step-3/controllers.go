package goose

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/eyeson-team/eyeson-go"
	"github.com/gofiber/fiber/v2"
)

func readUserFromCookie(c *fiber.Ctx) (*User, error) {
	id := c.Cookies("user", "0")
	if id == "0" {
		return nil, errors.New("No active user found")
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

func RootPath(c *fiber.Ctx) error {
	user, err := readUserFromCookie(c)
	if err != nil {
		return LoginsNew(c)
	}
	return WorkspacesIndex(c, user)
}

func LoginsNew(c *fiber.Ctx) error {
	return c.Render("logins/new", nil)
}

func LoginsCreate(c *fiber.Ctx) error {
	email := c.FormValue("email")
	// validate email
	user, err := NewUser(email)
	if err != nil {
		return fiber.ErrBadRequest
	}
	login, err := NewLogin(user)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err = login.SendMail(c.BaseURL()); err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	return c.Render("logins/instructions", nil)
}

func LoginsShow(c *fiber.Ctx) error {
	var login Login
	result := db.First(&login, "auth_key = ? AND expires_at >= ?", c.Params("auth"), time.Now())
	if result.Error != nil {
		log.Println("Invalid login attemp:", result.Error)
		return fiber.ErrBadRequest
	}
	log.Println("Login of User", login.UserID)
	c.Cookie(&fiber.Cookie{Name: "user", Value: fmt.Sprintf("%d", login.UserID)})
	db.Delete(&login)
	return c.Redirect("/", 303)
}

func Logout(c *fiber.Ctx) error {
	c.ClearCookie()
	return c.Redirect("/", 303)
}

func WorkspacesIndex(c *fiber.Ctx, user *User) error {
	var workspaces []Workspace
	db.Find(&workspaces)

	return c.Render("workspaces/index", fiber.Map{
		"User":       user,
		"Workspaces": workspaces,
	})
}

func WorkspacesCreate(c *fiber.Ctx, user *User) error {
	db.Create(&Workspace{Topic: c.FormValue("topic")})
	return c.Redirect("/", 303)
}

func WorkspacesShow(c *fiber.Ctx, user *User) error {
	var workspace *Workspace
	result := db.Preload("Meetings").Preload("Recordings").First(&workspace, "id = ?", c.Params("id"))
	if result.Error != nil {
		return fiber.ErrBadRequest
	}
	return c.Render("workspaces/show", fiber.Map{
		"User":      user,
		"Workspace": workspace,
	})
}

func MeetingsCreate(c *fiber.Ctx, user *User) error {
	id := c.Params("id")
	_, err := FindWorkspace(id)
	if err != nil {
		return fiber.ErrNotFound
	}
	options := map[string]string{
		"options[sfu_mode]": "disabled",
		"options[exit_url]": c.BaseURL() + "/",
	}
	// user, err := readUserFromCookie(c)
	// if err != nil {
	// 	return fiber.ErrBadRequest
	// }
	room, err := videoService.Rooms.Join(id, user.Name, options)
	if err != nil {
		log.Println(err)
		return fiber.ErrBadRequest
	}
	// return c.Redirect(room.Data.Links.Gui, 303)
	return c.Render("meetings/show", fiber.Map{"AccessKey": room.Data.AccessKey})
}

func WebhooksCreate(c *fiber.Ctx) error {
	// SKIPPED: validate webhook
	var webhook eyeson.Webhook
	if err := json.Unmarshal(c.Body(), &webhook); err != nil {
		log.Printf("Could not parse webhook request, %v.", err)
		return nil
	}
	log.Printf("Store new webhook %s for room %s\n", webhook.Type, webhook.Room.Id)
	if err := storeWebhookEvent(&webhook); err != nil {
		log.Printf("Failed to store webhook, %v.", err)
	}
	return nil
}

func storeWebhookEvent(webhook *eyeson.Webhook) error {
	switch webhook.Type {
	case "room_update":
		workspace, err := FindWorkspace(webhook.Room.Id)
		if err != nil {
			return err
		}
		if webhook.Room.Shutdown {
			meeting, err := workspace.LastMeeting()
			if err != nil {
				return err
			}
			return meeting.StoreShutdown()
		} else {
			_, err := workspace.NewMeeting(webhook.Room.StartedAt)
			return err
		}
	case "recording_update":
		workspace, err := FindWorkspace(webhook.Recording.Room.Id)
		if err != nil {
			return err
		}
		startedAt := time.Unix(int64(webhook.Recording.CreatedAt), 0)
		rec := &Recording{Workspace: *workspace, Reference: webhook.Recording.Id,
			StartedAt: startedAt, EndedAt: time.Now().UTC(),
			Duration: uint(webhook.Recording.Duration)}
		db.Create(rec)

		// if link is present, download recording to storage
		if len(webhook.Recording.Links.Download) > 0 {
			return rec.Store(webhook.Recording.Links.Download)
		}
	default:
		log.Println("Received unknown webhook type", webhook.Type)
	}
	return nil
}

func RecordingsShow(c *fiber.Ctx, user *User) error {
	var recording Recording
	result := db.First(&recording, "id = ?", c.Params("id"))
	if result.Error != nil {
		return fiber.ErrBadRequest
	}
	return c.SendFile(recording.StoragePath())
}
