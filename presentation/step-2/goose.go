package goose

import (
	"errors"

	"github.com/eyeson-team/eyeson-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var videoService *eyeson.Client

func RegisterRoutes(app *fiber.App) {
	app.Get("/", HandleRootPath)

	app.Post("/workspaces", authenticated(WorkspacesCreate))
	app.Get("/workspaces/:id", authenticated(WorkspacesShow))
	app.Post("/workspaces/:id/meeting", authenticated(MeetingsCreate))

	app.Post("/login", LoginsCreate)
	app.Get("/login/:auth", LoginsShow)
	app.Get("/logout", LoginsDelete)
}

func InitDatabase(filename string) error {
	var err error
	db, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Workspace{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Login{})
	return nil
}

func InitEyeson(apiKey string) error {
	if len(apiKey) == 0 {
		return errors.New("API_KEY not set")
	}
	videoService = eyeson.NewClient(apiKey)
	return nil
}
