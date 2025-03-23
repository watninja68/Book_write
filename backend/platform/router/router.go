// platform/router/router.go
package router

import (
	"encoding/gob"
	
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html"
	
	"01-Login/platform/authenticator"
	"01-Login/platform/middleware"
	"01-Login/web/app/callback"
	"01-Login/web/app/login"
	"01-Login/web/app/logout"
	"01-Login/web/app/user"
)

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator) *fiber.App {
	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})
	
	// Set up template engine
	viewEngine := html.New("./web/template", ".html")
	
	// Initialize Fiber
	app := fiber.New(fiber.Config{
		Views: viewEngine,
	})
	
	// Session store
	store := session.New()
	
	// Serve static files
	app.Static("/public", "web/static")
	
	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("home", fiber.Map{})
	})
	
	app.Get("/login", func(c *fiber.Ctx) error {
		return login.Handler(auth)(c)
	})
	
	app.Get("/callback", func(c *fiber.Ctx) error {
		return callback.Handler(auth)(c)
	})
	
	app.Get("/user", user.Handler)
	
	app.Get("/logout", logout.Handler)
	
	return app
}
