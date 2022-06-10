package goose

import (
	"errors"
	"strings"

	"github.com/eyeson-team/eyeson-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var videoService *eyeson.Client

func RegisterRoutes(app *fiber.App) {
	app.Get("/", RootPath)
	app.Post("/workspaces", authenticated(WorkspacesCreate))
	app.Get("/workspaces/:id", authenticated(WorkspacesShow))
	app.Post("/workspaces/:id/meeting", authenticated(MeetingsCreate))

	app.Get("/recordings/:id", authenticated(RecordingsShow))

	app.Post("/login", LoginsCreate)
	app.Get("/login/:auth", LoginsShow)
	app.Get("/logout", Logout)

	app.Post("/webhook", WebhooksCreate)
}

func InitDatabase(filename string) (err error) {
	db, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Workspace{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Login{})

	db.AutoMigrate(&Meeting{})
	db.AutoMigrate(&Recording{})
	return nil
}

func InitEyeson(apiKey string) error {
	if len(apiKey) == 0 {
		return errors.New("API_KEY not set")
	}
	videoService = eyeson.NewClient(apiKey)
	return nil
}

func RegisterWebhook(baseURL string) error {
	if len(baseURL) == 0 {
		return errors.New("WH_URL not set")
	}
	options := strings.Join([]string{
		eyeson.WEBHOOK_ROOM,
		eyeson.WEBHOOK_RECORDING,
	}, ",")
	return videoService.Webhook.Register(baseURL+"/webhook", options)
}
