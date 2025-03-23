// web/app/callback/callback.go
package callback

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	
	"01-Login/platform/authenticator"
)

// Handler for our callback.
func Handler(auth *authenticator.Authenticator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session from store
		store := session.New()
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		
		// Verify state parameter
		if c.Query("state") != sess.Get("state") {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid state parameter.")
		}
		
		// Exchange an authorization code for a token.
		token, err := auth.Exchange(c.Context(), c.Query("code"))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Failed to exchange an authorization code for a token.")
		}
		
		// Verify ID token
		idToken, err := auth.VerifyIDToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to verify ID Token.")
		}
		
		// Extract profile information from claims
		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		
		// Store tokens and profile in session
		sess.Set("access_token", token.AccessToken)
		sess.Set("profile", profile)
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		
		// Redirect to logged in page.
		return c.Redirect("/user", fiber.StatusTemporaryRedirect)
	}
}
