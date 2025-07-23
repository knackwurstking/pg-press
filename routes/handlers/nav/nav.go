package nav

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/internal/notifications"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
	feedNotifier     *notifications.FeedNotifier
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS, feedNotifier *notifications.FeedNotifier) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
		feedNotifier:     feedNotifier,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// This approach keeps the WebSocket handler within Echo's middleware chain
	// which is better for authentication that depends on Echo's context
	e.GET(h.serverPathPrefix+"/nav/feed-counter", h.handleFeedCounterWebSocketEcho)
}

// getUserFromWebSocket extracts user information from the WebSocket connection request
func (h *Handler) getUserFromWebSocket(ws *websocket.Conn) (*pgvis.User, error) {
	// Get the original HTTP request from the WebSocket connection
	req := ws.Request()
	if req == nil {
		return nil, fmt.Errorf("no HTTP request found in WebSocket connection")
	}

	// Create a temporary echo context to use existing authentication logic
	// This is a bit of a workaround since we need to adapt the echo-based auth
	// to work with the WebSocket connection

	// Try to get user from session/cookies in the request
	// We'll need to extract the authentication information from the request

	// Get cookies from the request
	cookies := req.Cookies()

	// Look for authentication cookies or headers
	// This depends on how your authentication is implemented
	// For now, we'll try to extract from cookies

	var apiKey string
	var sessionToken string

	for _, cookie := range cookies {
		switch cookie.Name {
		case "api_key":
			apiKey = cookie.Value
		case "session_token":
			sessionToken = cookie.Value
		}
	}

	// If we have an API key, try to authenticate with that
	if apiKey != "" {
		user, err := h.authenticateWithAPIKey(apiKey)
		if err == nil && user != nil {
			return user, nil
		}
	}

	// If we have a session token, try to authenticate with that
	if sessionToken != "" {
		user, err := h.authenticateWithSession(sessionToken)
		if err == nil && user != nil {
			return user, nil
		}
	}

	return nil, fmt.Errorf("authentication failed: no valid credentials found")
}

// authenticateWithAPIKey authenticates a user using an API key
func (h *Handler) authenticateWithAPIKey(apiKey string) (*pgvis.User, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("empty API key")
	}

	// Use the database to find user by API key
	user, err := h.db.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by API key: %w", err)
	}

	return user, nil
}

// authenticateWithSession authenticates a user using a session token
func (h *Handler) authenticateWithSession(sessionToken string) (*pgvis.User, error) {
	if sessionToken == "" {
		return nil, fmt.Errorf("empty session token")
	}

	// This would depend on your session implementation
	// For now, we'll return an error since session auth isn't implemented in the provided context
	return nil, fmt.Errorf("session authentication not implemented")
}

// Alternative approach: Create an echo-compatible WebSocket handler
func (h *Handler) handleFeedCounterWebSocketEcho(c echo.Context) error {
	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		// Get user from echo context
		user, herr := utils.GetUserFromContext(c)
		if herr != nil {
			log.Printf("[WebSocket] User authentication failed: %v", herr)
			ws.Close()
			return
		}

		log.Printf("[WebSocket] User authenticated: %s (LastFeed: %d)", user.UserName, user.LastFeed)

		// Register the connection with the feed notifier
		feedConn := h.feedNotifier.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			log.Printf("[WebSocket] Failed to register connection for user %s", user.UserName)
			ws.Close()
			return
		}

		log.Printf("[WebSocket] Connection registered for user %s", user.UserName)

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(h.feedNotifier)

		log.Printf("[WebSocket] Connection closed for user %s", user.UserName)
	})

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
