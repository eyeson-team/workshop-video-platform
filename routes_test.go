package goose

import "testing"

func TestRoutes(t *testing.T) {
	setupTestDB()
	defer removeTestDB()

	// req := httptest.NewRequest("GET", "/", nil)

	// AddRoutes(app)
	// resp, _ := app.Test(req)
	// if resp.StatusCode != fiber.StatusOK {
	// 	tkk
	// }
}
