package goose

import (
	"github.com/gofiber/fiber/v2"
)

func WorkspacesIndex(c *fiber.Ctx) error {
	var workspaces []Workspace
	db.Find(&workspaces)

	return c.Render("workspaces/index", fiber.Map{"Workspaces": workspaces})
}

func WorkspacesCreate(c *fiber.Ctx) error {
	// SKIPPED: validate topic min/max lenght.
	db.Create(&Workspace{Topic: c.FormValue("topic")})
	return c.Redirect("/", 303)
}

func WorkspacesShow(c *fiber.Ctx) error {
	var workspace Workspace
	result := db.First(&workspace, "id = ?", c.Params("id"))
	if result.Error != nil {
		return fiber.ErrBadRequest
	}
	return c.Render("workspaces/show", fiber.Map{"Workspace": workspace})
}

func MeetingsCreate(c *fiber.Ctx) error {
	// ...
}
