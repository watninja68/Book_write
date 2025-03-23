// web/app/logout/logout.go
package logout

import (
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

// Handler for our logout.
func Handler(c *fiber.Ctx) error {
	// Build Auth0 logout URL
	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Determine protocol (http or https)
	scheme := "http"
	if c.Protocol() == "https" {
		scheme = "https"
	}

	// Build the return URL
	returnTo, err := url.Parse(scheme + "://" + c.Hostname())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Add query parameters
	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	// Redirect to Auth0 logout page
	return c.Redirect(logoutUrl.String(), fiber.StatusTemporaryRedirect)
}
