package goose

import (
	"errors"

	"github.com/eyeson-team/eyeson-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDatabase does open a connection to our sqlite database and migrate our
// models (tables).
func InitDatabase(filename string) error {
	var err error
	db, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Workspace{})
	return nil
}

// RegisterRoutes does register routes with naming convention ressouces#action.
func RegisterRoutes(app *fiber.App) {
	// WorkspacesAction
	app.Get("/", WorkspacesIndex)
	app.Post("/workspaces", WorkspacesCreate)
	app.Get("/workspaces/:id", WorkspacesShow)
	app.Post("/workspaces/:workspace_id/meeting", MeetingsCreate)
}

var videoService *eyeson.Client

// InitEyeson does initialize the video service with the given eyeson API key.
func InitEyeson(apiKey string) error {
	if len(apiKey) == 0 {
		return errors.New("API_KEY not set")
	}
	videoService = eyeson.NewClient(apiKey)
	return nil
}
