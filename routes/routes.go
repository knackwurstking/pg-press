// Package routes provides HTTP route handlers and web interface for the pgvis application.
//
// This package implements the web interface layer, handling HTTP requests and responses,
// template rendering, authentication, and session management. It serves as the main
// entry point for user interactions with the system.
//
// Key Components:
//   - Authentication and session management via cookies
//   - Template rendering for HTML pages
//   - Static file serving
//   - Integration with various feature modules (feed, profile, trouble reports)
//
// The package uses Echo framework for HTTP routing and embedded file systems
// for templates and static assets.
package routes

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/feed"
	"github.com/knackwurstking/pg-vis/routes/nav"
	"github.com/knackwurstking/pg-vis/routes/profile"
	"github.com/knackwurstking/pg-vis/routes/troublereports"
)

const (
	// CookieName is the name used for authentication cookies
	CookieName = "pgvis-api-key"

	// CookieExpirationDuration defines how long authentication cookies remain valid
	// Set to 6 months (24 hours * 31 days * 6 months)
	CookieExpirationDuration = time.Hour * 24 * 31 * 6
)

// Embedded file systems for templates and static assets
var (
	//go:embed templates
	templates embed.FS

	//go:embed static
	static embed.FS
)

// Options contains configuration options for setting up HTTP routes.
type Options struct {
	// ServerPathPrefix is the URL path prefix for all routes (e.g., "/app")
	ServerPathPrefix string

	// DB provides access to the database layer
	DB *pgvis.DB
}

// Serve sets up all HTTP routes and handlers for the pgvis web application.
// It configures static file serving, authentication routes, and delegates
// to specialized route handlers for different features.
//
// Parameters:
//   - e: Echo instance to register routes with
//   - o: Configuration options including path prefix and database access
func Serve(e *echo.Echo, o Options) {
	// Serve static files (CSS, JS, images, etc.)
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	// Core application routes
	serveHome(e, o)
	serveLogin(e, o)
	serveLogout(e, o)

	// Feature-specific route modules
	feed.Serve(templates, o.ServerPathPrefix, e, o.DB)
	profile.Serve(templates, o.ServerPathPrefix, e, o.DB)
	troublereports.Serve(templates, o.ServerPathPrefix, e, o.DB)
	nav.Serve(templates, o.ServerPathPrefix, e, o.DB)
}

// serveHome handles the home page route, rendering the main application interface.
func serveHome(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		// Parse required templates for the home page
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/home.html",
			"templates/nav/feed.html",
		)
		if err != nil {
			log.Errorf("Failed to parse home page templates: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to load page templates")
		}

		// Render the home page
		if err := t.Execute(c.Response(), nil); err != nil {
			log.Errorf("Failed to execute home page template: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to render page")
		}

		return nil
	})
}

// LoginPageData contains data passed to the login page template.
type LoginPageData struct {
	// ApiKey is the API key submitted by the user (if any)
	ApiKey string

	// InvalidApiKey indicates whether the submitted API key was invalid
	InvalidApiKey bool
}

// serveLogin handles the login page route, processing API key authentication.
func serveLogin(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/login", func(c echo.Context) error {
		// Get form parameters
		formParams, _ := c.FormParams()
		apiKey := formParams.Get("api-key")

		// Attempt login if API key is provided
		if apiKey != "" {
			if ok, err := handleApiKeyLogin(apiKey, options.DB, c); ok {
				// Successful login - redirect to profile
				if err := c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
					log.Errorf("Failed to redirect after login: %v", err)
					return echo.NewHTTPError(http.StatusInternalServerError,
						"Redirect failed")
				}
				return nil
			} else if err != nil {
				// Log authentication errors
				log.Warnf("Login failed for API key: %v", err)
			}
		}

		// Prepare page data
		pageData := LoginPageData{
			ApiKey:        apiKey,
			InvalidApiKey: apiKey != "",
		}

		// Parse and execute login template
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/login.html",
		)
		if err != nil {
			log.Errorf("Failed to parse login page templates: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to load page templates")
		}

		if err := t.Execute(c.Response(), pageData); err != nil {
			log.Errorf("Failed to execute login page template: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to render page")
		}

		return nil
	})
}

// serveLogout handles the logout route, clearing authentication cookies and sessions.
func serveLogout(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/logout", func(c echo.Context) error {
		// Get the current authentication cookie
		if cookie, err := c.Cookie(CookieName); err == nil {
			// Remove the session from database
			if err := options.DB.Cookies.Remove(cookie.Value); err != nil {
				log.Warnf("Failed to remove cookie from database: %v", err)
				// Continue with logout even if database removal fails
			}
		}

		// Redirect to login page
		if err := c.Redirect(http.StatusSeeOther, "./login"); err != nil {
			log.Errorf("Failed to redirect after logout: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Redirect failed")
		}

		return nil
	})
}

// handleApiKeyLogin processes API key authentication and creates a session.
// It validates the API key, creates a new session cookie, and stores session data.
//
// Parameters:
//   - apiKey: The API key to authenticate
//   - db: Database connection for user and session operations
//   - ctx: Echo context for cookie management
//
// Returns:
//   - ok: true if authentication was successful
//   - err: error if authentication failed due to system issues
func handleApiKeyLogin(apiKey string, db *pgvis.DB, ctx echo.Context) (ok bool, err error) {
	if apiKey == "" {
		return false, nil
	}

	// Validate API key and get user
	user, err := db.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if pgvis.IsNotFound(err) {
			return false, nil // Invalid API key, but not a system error
		}
		return false, fmt.Errorf("database error during authentication: %w", err)
	}

	// Verify API key matches (additional security check)
	if user.ApiKey != apiKey {
		return false, nil
	}

	// Remove any existing cookie for this user
	if existingCookie, err := ctx.Cookie(CookieName); err == nil {
		log.Debug("Removing existing authentication cookie")
		if err := db.Cookies.Remove(existingCookie.Value); err != nil {
			log.Warnf("Failed to remove existing cookie: %v", err)
			// Continue with login process even if cleanup fails
		}
	}

	log.Debugf("Creating new session for user %s (Telegram ID: %d)",
		user.UserName, user.TelegramID)

	// Create new session cookie
	cookie := &http.Cookie{
		Name:    CookieName,
		Value:   uuid.New().String(),
		Expires: time.Now().Add(CookieExpirationDuration),
	}

	ctx.SetCookie(cookie)

	// Store session in database
	session := &pgvis.Cookie{
		UserAgent: ctx.Request().UserAgent(),
		Value:     cookie.Value,
		ApiKey:    apiKey,
		LastLogin: time.Now().UnixMilli(),
	}

	if err := db.Cookies.Add(session); err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}

	return true, nil
}
