package goose

import (
	"github.com/gofiber/fiber/v2"
)

// AddRoutes does register all HTTP server routes to the fiber app.
//
// As HTML forms do not support HTTP methods other than POST, we stick to POST
// for the update routes.
//
// All controller functions can be found in `./controllers.go`.
func AddRoutes(app *fiber.App) {
	app.Get("/", HandleRootPath)

	app.Post("/sessions", LoginsCreate)
	app.Get("/sessions/:auth", LoginsShow)

	app.Get("/workspaces/new", authenticated(WorkspacesNew))
	app.Post("/workspaces", authenticated(WorkspacesCreate))
	app.Get("/workspaces/:id", authenticated(WorkspacesShow))
	app.Get("/workspaces/:id/edit", authenticated(WorkspacesEdit))
	app.Post("/workspaces/:id/edit", authenticated(WorkspacesUpdate))

	app.Post("/workspaces/:id/meeting", authenticated(MeetingsCreate))
	app.Post("/workspaces/:id/users", authenticated(WorkspacesAddUser))

	app.Get("/recordings/:id.webm", authenticated(RecordingsShow))
	app.Get("/snapshots/:id.jpg", authenticated(SnapshotsShow))

	app.Get("/meetings/:id/edit", authenticated(MeetingsEdit))
	app.Post("/meetings/:id/edit", authenticated(MeetingsUpdate))

	// Provide an endpoint for incoming webhooks from the eyeson service.
	app.Post("/webhook", WebhooksCreate)

	// signout clears the cookie and redirects to the root path.
	app.Get("/signout", func(c *fiber.Ctx) error {
		c.ClearCookie()
		return c.Redirect("/", 303)
	})

	// Add static assets from local directory.
	app.Static("/assets/", "./assets")

	// Add a debugging route for local requests.
	app.Get("/stack", func(c *fiber.Ctx) error {
		if c.IsFromLocal() == false {
			return fiber.ErrNotFound
		}
		return c.JSON(c.App().Stack())
	})
}
