// platform/middleware/isAuthenticated.go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(c *fiber.Ctx) error {
	// Get session from store
	store := session.New()
	sess, err := store.Get(c)
	if err != nil {
		return c.Redirect("/", fiber.StatusSeeOther)
	}

	// Check if user is authenticated
	if sess.Get("profile") == nil {
		return c.Redirect("/", fiber.StatusSeeOther)
	}

	// User is authenticated, continue to next middleware/handler
	return c.Next()
}
