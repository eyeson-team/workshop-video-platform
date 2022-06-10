package main

import (
	"flag"
	"fmt"
	"goose"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

func main() {
	port := flag.Int("port", 8077, "server listening port")
	flag.Parse()

	if err := goose.InitDatabase("local.db"); err != nil {
		log.Fatal(err)
	}
	if err := goose.InitEyeson(os.Getenv("API_KEY")); err != nil {
		log.Fatal(err)
	}

	engine := html.New("./views", ".tmpl")
	app := fiber.New(fiber.Config{Views: engine, ViewsLayout: "layouts/main"})
	// app.Get("/", func(c *fiber.Ctx) error {
	// 	return c.SendString("hello world\n")
	// })
	goose.RegisterRoutes(app)

	app.Static("/assets/", "./assets")

	fmt.Printf("Start server on port %d\n", *port)
	if err := app.Listen(fmt.Sprintf(":%d", *port)); err != nil {
		log.Fatal(err)
	}
}
