// web/app/user/user.go
package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Handler for our logged-in user page.
func Handler(c *fiber.Ctx) error {
	// Get session from store
	store := session.New()
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	
	// Get profile from session
	profile := sess.Get("profile")
	
	// Render user template with profile data
	return c.Render("user", fiber.Map{
		"profile": profile,
	})
}
