package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goose"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/favicon"
)

// main does handle all setup and teardown for our web application. The
// platform uses the [fiber framework](https://docs.gofiber.io) and the
// database orm [gorm](https://gorm.io/docs/).
func main() {
	port := flag.Int("port", 8077, "server listening port")
	flag.Parse()

	engine := goose.NewViewEngine("./views")
	app := fiber.New(fiber.Config{Views: engine, ViewsLayout: "layouts/main"})

	// Connect and migrate the sqlite database on a predefined static path.
	if err := goose.InitDatabase("./db/production.db"); err != nil {
		log.Fatal("Failed to connect database: %s.", err)
	}
	// Configure eyeson API to join video meetings on-the-fly.
	if err := goose.InitEyeson(os.Getenv("API_KEY")); err != nil {
		log.Fatal("eyeson API_KEY not found: %s.", err)
	}
	// Register the webhook endpoint at eyeson such that we can receive any
	// meeting and recording updates.
	if err := goose.RegisterWebhook(os.Getenv("WH_URL")); err != nil {
		log.Fatalf("Webhook could not be registered: %s.", err)
	}

	// Read cookie secret or generate a new one on-the-fly. A non-persistent
	// secret will invalidate cookies with every restart of the application.
	key := os.Getenv("COOKIE_SECRET")
	if len(key) == 0 {
		key = encryptcookie.GenerateKey()
		log.Println("No cookie secret has been set, generating a new one:", key)
	}
	app.Use(encryptcookie.New(encryptcookie.Config{Key: key}))

	// Use favicon middleware to serve favicon fast.
	app.Use(favicon.New(favicon.Config{File: "./assets/favicon.ico"}))

	// Register all routes from the video platform application.
	goose.AddRoutes(app)

	// Handle server shutdown gracefully such that cleanup procedures can be
	// executed properly.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println("Start server on port", *port)
		app.Listen(fmt.Sprintf(":%d", *port))
	}()
	<-done

	log.Println("Shutting down server...")
	// goose.UnregisterWebhook()
	// log.Println("Webhook unregistered")
}
